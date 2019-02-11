// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0
//
//
// How to manually run individual tests against the real cloud
//
// $ go test -v git.curoverse.com/arvados.git/lib/cloud/azure -live-azure-cfg azconfig.yml -check.f=TestListInstances
//
// Example azconfig.yml:
//
// ImageIDForTestSuite: "https://example.blob.core.windows.net/system/Microsoft.Compute/Images/images/zzzzz-compute-osDisk.XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX.vhd"
// SubscriptionID: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
// ClientID: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
// Location: centralus
// CloudEnvironment: AzurePublicCloud
// ClientSecret: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
// TenantId: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
// ResourceGroup: zzzzz
// Network: zzzzz0:10 / 3:26:1
// Subnet: zzzzz-subnet-private
// StorageAccount: example
// BlobContainer: vhds
// DeleteDanglingResourcesAfter: 20s

package azure

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"git.curoverse.com/arvados.git/lib/cloud"
	"git.curoverse.com/arvados.git/sdk/go/arvados"
	"git.curoverse.com/arvados.git/sdk/go/config"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-06-01/network"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/jmcvetta/randutil"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	check "gopkg.in/check.v1"
)

// Gocheck boilerplate
func Test(t *testing.T) {
	check.TestingT(t)
}

type AzureInstanceSetSuite struct{}

var _ = check.Suite(&AzureInstanceSetSuite{})

type VirtualMachinesClientStub struct{}

var testKey = []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDLQS1ExT2+WjA0d/hntEAyAtgeN1W2ik2QX8c2zO6HjlPHWXL92r07W0WMuDib40Pcevpi1BXeBWXA9ZB5KKMJB+ukaAu22KklnQuUmNvk6ZXnPKSkGxuCYvPQb08WhHf3p1VxiKfP3iauedBDM4x9/bkJohlBBQiFXzNUcQ+a6rKiMzmJN2gbL8ncyUzc+XQ5q4JndTwTGtOlzDiGOc9O4z5Dd76wtAVJneOuuNpwfFRVHThpJM6VThpCZOnl8APaceWXKeuwOuCae3COZMz++xQfxOfZ9Z8aIwo+TlQhsRaNfZ4Vjrop6ej8dtfZtgUFKfbXEOYaHrGrWGotFDTD example@example`)

func (*VirtualMachinesClientStub) createOrUpdate(ctx context.Context,
	resourceGroupName string,
	VMName string,
	parameters compute.VirtualMachine) (result compute.VirtualMachine, err error) {
	parameters.ID = &VMName
	parameters.Name = &VMName
	return parameters, nil
}

func (*VirtualMachinesClientStub) delete(ctx context.Context, resourceGroupName string, VMName string) (result *http.Response, err error) {
	return nil, nil
}

func (*VirtualMachinesClientStub) listComplete(ctx context.Context, resourceGroupName string) (result compute.VirtualMachineListResultIterator, err error) {
	return compute.VirtualMachineListResultIterator{}, nil
}

type InterfacesClientStub struct{}

func (*InterfacesClientStub) createOrUpdate(ctx context.Context,
	resourceGroupName string,
	nicName string,
	parameters network.Interface) (result network.Interface, err error) {
	parameters.ID = to.StringPtr(nicName)
	(*parameters.IPConfigurations)[0].PrivateIPAddress = to.StringPtr("192.168.5.5")
	return parameters, nil
}

func (*InterfacesClientStub) delete(ctx context.Context, resourceGroupName string, VMName string) (result *http.Response, err error) {
	return nil, nil
}

func (*InterfacesClientStub) listComplete(ctx context.Context, resourceGroupName string) (result network.InterfaceListResultIterator, err error) {
	return network.InterfaceListResultIterator{}, nil
}

var live = flag.String("live-azure-cfg", "", "Test with real azure API, provide config file")

func GetInstanceSet() (cloud.InstanceSet, cloud.ImageID, arvados.Cluster, error) {
	cluster := arvados.Cluster{
		InstanceTypes: arvados.InstanceTypeMap(map[string]arvados.InstanceType{
			"tiny": arvados.InstanceType{
				Name:         "tiny",
				ProviderType: "Standard_D1_v2",
				VCPUs:        1,
				RAM:          4000000000,
				Scratch:      10000000000,
				Price:        .02,
				Preemptible:  false,
			},
		})}
	if *live != "" {
		exampleCfg := make(map[string]interface{})
		err := config.LoadFile(&exampleCfg, *live)
		if err != nil {
			return nil, cloud.ImageID(""), cluster, err
		}
		ap, err := newAzureInstanceSet(exampleCfg, "test123", logrus.StandardLogger())
		return ap, cloud.ImageID(exampleCfg["ImageIDForTestSuite"].(string)), cluster, err
	}
	ap := azureInstanceSet{
		azconfig: azureInstanceSetConfig{
			BlobContainer: "vhds",
		},
		dispatcherID: "test123",
		namePrefix:   "compute-test123-",
		logger:       logrus.StandardLogger(),
		deleteNIC:    make(chan string),
		deleteBlob:   make(chan storage.Blob),
	}
	ap.ctx, ap.stopFunc = context.WithCancel(context.Background())
	ap.vmClient = &VirtualMachinesClientStub{}
	ap.netClient = &InterfacesClientStub{}
	return &ap, cloud.ImageID("blob"), cluster, nil
}

func (*AzureInstanceSetSuite) TestCreate(c *check.C) {
	ap, img, cluster, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey(testKey)
	c.Assert(err, check.IsNil)

	nodetoken, err := randutil.String(40, "abcdefghijklmnopqrstuvwxyz0123456789")
	c.Assert(err, check.IsNil)

	inst, err := ap.Create(cluster.InstanceTypes["tiny"],
		img, map[string]string{
			"node-token": nodetoken},
		pk)

	c.Assert(err, check.IsNil)

	tg := inst.Tags()
	log.Printf("Result %v %v %v", inst.String(), inst.Address(), tg)

}

func (*AzureInstanceSetSuite) TestListInstances(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}

	l, err := ap.Instances(nil)

	c.Assert(err, check.IsNil)

	for _, i := range l {
		tg := i.Tags()
		log.Printf("%v %v %v", i.String(), i.Address(), tg)
	}
}

func (*AzureInstanceSetSuite) TestManageNics(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}

	ap.(*azureInstanceSet).manageNics()
	ap.Stop()
}

func (*AzureInstanceSetSuite) TestManageBlobs(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}

	ap.(*azureInstanceSet).manageBlobs()
	ap.Stop()
}

func (*AzureInstanceSetSuite) TestDestroyInstances(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}

	l, err := ap.Instances(nil)
	c.Assert(err, check.IsNil)

	for _, i := range l {
		c.Check(i.Destroy(), check.IsNil)
	}
}

func (*AzureInstanceSetSuite) TestDeleteFake(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}

	_, err = ap.(*azureInstanceSet).netClient.delete(context.Background(), "fakefakefake", "fakefakefake")

	de, ok := err.(autorest.DetailedError)
	if ok {
		rq := de.Original.(*azure.RequestError)

		log.Printf("%v %q %q", rq.Response.StatusCode, rq.ServiceError.Code, rq.ServiceError.Message)
	}
}

func (*AzureInstanceSetSuite) TestWrapError(c *check.C) {
	retryError := autorest.DetailedError{
		Original: &azure.RequestError{
			DetailedError: autorest.DetailedError{
				Response: &http.Response{
					StatusCode: 429,
					Header:     map[string][]string{"Retry-After": []string{"123"}},
				},
			},
			ServiceError: &azure.ServiceError{},
		},
	}
	wrapped := wrapAzureError(retryError)
	_, ok := wrapped.(cloud.RateLimitError)
	c.Check(ok, check.Equals, true)

	quotaError := autorest.DetailedError{
		Original: &azure.RequestError{
			DetailedError: autorest.DetailedError{
				Response: &http.Response{
					StatusCode: 503,
				},
			},
			ServiceError: &azure.ServiceError{
				Message: "No more quota",
			},
		},
	}
	wrapped = wrapAzureError(quotaError)
	_, ok = wrapped.(cloud.QuotaError)
	c.Check(ok, check.Equals, true)
}

func (*AzureInstanceSetSuite) TestSetTags(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}
	l, err := ap.Instances(nil)
	c.Assert(err, check.IsNil)

	if len(l) > 0 {
		err = l[0].SetTags(map[string]string{"foo": "bar"})
		if err != nil {
			c.Fatal("Error setting tags", err)
		}
	}
	l, err = ap.Instances(nil)
	c.Assert(err, check.IsNil)

	if len(l) > 0 {
		tg := l[0].Tags()
		log.Printf("tags are %v", tg)
	}
}

func (*AzureInstanceSetSuite) TestSSH(c *check.C) {
	ap, _, _, err := GetInstanceSet()
	if err != nil {
		c.Fatal("Error making provider", err)
	}
	l, err := ap.Instances(nil)
	c.Assert(err, check.IsNil)

	if len(l) > 0 {

		sshclient, err := SetupSSHClient(c, l[0])
		c.Assert(err, check.IsNil)

		sess, err := sshclient.NewSession()
		c.Assert(err, check.IsNil)

		out, err := sess.Output("cat /home/crunch/node-token")
		c.Assert(err, check.IsNil)

		log.Printf("%v", string(out))

		sshclient.Conn.Close()
	}
}

func SetupSSHClient(c *check.C, inst cloud.Instance) (*ssh.Client, error) {
	addr := inst.Address() + ":2222"
	if addr == "" {
		return nil, errors.New("instance has no address")
	}

	f, err := os.Open("azconfig_sshkey")
	c.Assert(err, check.IsNil)

	keybytes, err := ioutil.ReadAll(f)
	c.Assert(err, check.IsNil)

	priv, err := ssh.ParsePrivateKey(keybytes)
	c.Assert(err, check.IsNil)

	var receivedKey ssh.PublicKey
	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: "crunch",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(priv),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			receivedKey = key
			return nil
		},
		Timeout: time.Minute,
	})

	if err != nil {
		return nil, err
	} else if receivedKey == nil {
		return nil, errors.New("BUG: key was never provided to HostKeyCallback")
	}

	err = inst.VerifyHostKey(receivedKey, client)
	c.Assert(err, check.IsNil)

	return client, nil
}
