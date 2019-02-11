// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.curoverse.com/arvados.git/lib/cloud"
	"git.curoverse.com/arvados.git/sdk/go/arvados"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-06-01/network"
	storageacct "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-02-01/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/jmcvetta/randutil"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Driver provides access to the azure instance set
var Driver = cloud.DriverFunc(newAzureInstanceSet)

type azureInstanceSetConfig struct {
	SubscriptionID               string
	ClientID                     string
	ClientSecret                 string
	TenantID                     string
	CloudEnvironment             string
	ResourceGroup                string
	Location                     string
	Network                      string
	Subnet                       string
	StorageAccount               string
	BlobContainer                string
	DeleteDanglingResourcesAfter arvados.Duration
}

type virtualMachinesClientWrapper interface {
	createOrUpdate(ctx context.Context,
		resourceGroupName string,
		VMName string,
		parameters compute.VirtualMachine) (result compute.VirtualMachine, err error)
	delete(ctx context.Context, resourceGroupName string, VMName string) (result *http.Response, err error)
	listComplete(ctx context.Context, resourceGroupName string) (result compute.VirtualMachineListResultIterator, err error)
}

type virtualMachinesClientImpl struct {
	inner compute.VirtualMachinesClient
}

func (cl *virtualMachinesClientImpl) createOrUpdate(ctx context.Context,
	resourceGroupName string,
	VMName string,
	parameters compute.VirtualMachine) (result compute.VirtualMachine, err error) {

	future, err := cl.inner.CreateOrUpdate(ctx, resourceGroupName, VMName, parameters)
	if err != nil {
		return compute.VirtualMachine{}, wrapAzureError(err)
	}
	future.WaitForCompletionRef(ctx, cl.inner.Client)
	r, err := future.Result(cl.inner)
	return r, wrapAzureError(err)
}

func (cl *virtualMachinesClientImpl) delete(ctx context.Context, resourceGroupName string, VMName string) (result *http.Response, err error) {
	future, err := cl.inner.Delete(ctx, resourceGroupName, VMName)
	if err != nil {
		return nil, wrapAzureError(err)
	}
	err = future.WaitForCompletionRef(ctx, cl.inner.Client)
	return future.Response(), wrapAzureError(err)
}

func (cl *virtualMachinesClientImpl) listComplete(ctx context.Context, resourceGroupName string) (result compute.VirtualMachineListResultIterator, err error) {
	r, err := cl.inner.ListComplete(ctx, resourceGroupName)
	return r, wrapAzureError(err)
}

type interfacesClientWrapper interface {
	createOrUpdate(ctx context.Context,
		resourceGroupName string,
		networkInterfaceName string,
		parameters network.Interface) (result network.Interface, err error)
	delete(ctx context.Context, resourceGroupName string, networkInterfaceName string) (result *http.Response, err error)
	listComplete(ctx context.Context, resourceGroupName string) (result network.InterfaceListResultIterator, err error)
}

type interfacesClientImpl struct {
	inner network.InterfacesClient
}

func (cl *interfacesClientImpl) delete(ctx context.Context, resourceGroupName string, VMName string) (result *http.Response, err error) {
	future, err := cl.inner.Delete(ctx, resourceGroupName, VMName)
	if err != nil {
		return nil, wrapAzureError(err)
	}
	err = future.WaitForCompletionRef(ctx, cl.inner.Client)
	return future.Response(), wrapAzureError(err)
}

func (cl *interfacesClientImpl) createOrUpdate(ctx context.Context,
	resourceGroupName string,
	networkInterfaceName string,
	parameters network.Interface) (result network.Interface, err error) {

	future, err := cl.inner.CreateOrUpdate(ctx, resourceGroupName, networkInterfaceName, parameters)
	if err != nil {
		return network.Interface{}, wrapAzureError(err)
	}
	future.WaitForCompletionRef(ctx, cl.inner.Client)
	r, err := future.Result(cl.inner)
	return r, wrapAzureError(err)
}

func (cl *interfacesClientImpl) listComplete(ctx context.Context, resourceGroupName string) (result network.InterfaceListResultIterator, err error) {
	r, err := cl.inner.ListComplete(ctx, resourceGroupName)
	return r, wrapAzureError(err)
}

var quotaRe = regexp.MustCompile(`(?i:exceed|quota|limit)`)

type azureRateLimitError struct {
	azure.RequestError
	firstRetry time.Time
}

func (ar *azureRateLimitError) EarliestRetry() time.Time {
	return ar.firstRetry
}

type azureQuotaError struct {
	azure.RequestError
}

func (ar *azureQuotaError) IsQuotaError() bool {
	return true
}

func wrapAzureError(err error) error {
	de, ok := err.(autorest.DetailedError)
	if !ok {
		return err
	}
	rq, ok := de.Original.(*azure.RequestError)
	if !ok {
		return err
	}
	if rq.Response == nil {
		return err
	}
	if rq.Response.StatusCode == 429 || len(rq.Response.Header["Retry-After"]) >= 1 {
		// API throttling
		ra := rq.Response.Header["Retry-After"][0]
		earliestRetry, parseErr := http.ParseTime(ra)
		if parseErr != nil {
			// Could not parse as a timestamp, must be number of seconds
			dur, parseErr := strconv.ParseInt(ra, 10, 64)
			if parseErr == nil {
				earliestRetry = time.Now().Add(time.Duration(dur) * time.Second)
			} else {
				// Couldn't make sense of retry-after,
				// so set retry to 20 seconds
				earliestRetry = time.Now().Add(20 * time.Second)
			}
		}
		return &azureRateLimitError{*rq, earliestRetry}
	}
	if rq.ServiceError == nil {
		return err
	}
	if quotaRe.FindString(rq.ServiceError.Code) != "" || quotaRe.FindString(rq.ServiceError.Message) != "" {
		return &azureQuotaError{*rq}
	}
	return err
}

type azureInstanceSet struct {
	azconfig          azureInstanceSetConfig
	vmClient          virtualMachinesClientWrapper
	netClient         interfacesClientWrapper
	storageAcctClient storageacct.AccountsClient
	azureEnv          azure.Environment
	interfaces        map[string]network.Interface
	dispatcherID      string
	namePrefix        string
	ctx               context.Context
	stopFunc          context.CancelFunc
	stopWg            sync.WaitGroup
	deleteNIC         chan string
	deleteBlob        chan storage.Blob
	logger            logrus.FieldLogger
}

func newAzureInstanceSet(config map[string]interface{}, dispatcherID cloud.InstanceSetID, logger logrus.FieldLogger) (prv cloud.InstanceSet, err error) {
	azcfg := azureInstanceSetConfig{}

	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: arvados.DurationMapStructureDecodeHook(),
		Result:     &azcfg}

	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		return nil, err
	}
	if err = decoder.Decode(config); err != nil {
		return nil, err
	}

	ap := azureInstanceSet{logger: logger}
	err = ap.setup(azcfg, string(dispatcherID))
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

func (az *azureInstanceSet) setup(azcfg azureInstanceSetConfig, dispatcherID string) (err error) {
	az.azconfig = azcfg
	vmClient := compute.NewVirtualMachinesClient(az.azconfig.SubscriptionID)
	netClient := network.NewInterfacesClient(az.azconfig.SubscriptionID)
	storageAcctClient := storageacct.NewAccountsClient(az.azconfig.SubscriptionID)

	az.azureEnv, err = azure.EnvironmentFromName(az.azconfig.CloudEnvironment)
	if err != nil {
		return err
	}

	authorizer, err := auth.ClientCredentialsConfig{
		ClientID:     az.azconfig.ClientID,
		ClientSecret: az.azconfig.ClientSecret,
		TenantID:     az.azconfig.TenantID,
		Resource:     az.azureEnv.ResourceManagerEndpoint,
		AADEndpoint:  az.azureEnv.ActiveDirectoryEndpoint,
	}.Authorizer()
	if err != nil {
		return err
	}

	vmClient.Authorizer = authorizer
	netClient.Authorizer = authorizer
	storageAcctClient.Authorizer = authorizer

	az.vmClient = &virtualMachinesClientImpl{vmClient}
	az.netClient = &interfacesClientImpl{netClient}
	az.storageAcctClient = storageAcctClient

	az.dispatcherID = dispatcherID
	az.namePrefix = fmt.Sprintf("compute-%s-", az.dispatcherID)

	az.ctx, az.stopFunc = context.WithCancel(context.Background())
	go func() {
		az.stopWg.Add(1)
		defer az.stopWg.Done()

		tk := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-az.ctx.Done():
				tk.Stop()
				return
			case <-tk.C:
				az.manageBlobs()
			}
		}
	}()

	az.deleteNIC = make(chan string)
	az.deleteBlob = make(chan storage.Blob)

	for i := 0; i < 4; i++ {
		go func() {
			for {
				nicname, ok := <-az.deleteNIC
				if !ok {
					return
				}
				_, delerr := az.netClient.delete(context.Background(), az.azconfig.ResourceGroup, nicname)
				if delerr != nil {
					az.logger.WithError(delerr).Warnf("Error deleting %v", nicname)
				} else {
					az.logger.Printf("Deleted NIC %v", nicname)
				}
			}
		}()
		go func() {
			for {
				blob, ok := <-az.deleteBlob
				if !ok {
					return
				}
				err := blob.Delete(nil)
				if err != nil {
					az.logger.WithError(err).Warnf("Error deleting %v", blob.Name)
				} else {
					az.logger.Printf("Deleted blob %v", blob.Name)
				}
			}
		}()
	}

	return nil
}

func (az *azureInstanceSet) Create(
	instanceType arvados.InstanceType,
	imageID cloud.ImageID,
	newTags cloud.InstanceTags,
	publicKey ssh.PublicKey) (cloud.Instance, error) {

	az.stopWg.Add(1)
	defer az.stopWg.Done()

	if len(newTags["node-token"]) == 0 {
		return nil, fmt.Errorf("Must provide tag 'node-token'")
	}

	name, err := randutil.String(15, "abcdefghijklmnopqrstuvwxyz0123456789")
	if err != nil {
		return nil, err
	}

	name = az.namePrefix + name

	timestamp := time.Now().Format(time.RFC3339Nano)

	tags := make(map[string]*string)
	tags["created-at"] = &timestamp
	for k, v := range newTags {
		newstr := v
		tags["dispatch-"+k] = &newstr
	}

	tags["dispatch-instance-type"] = &instanceType.Name

	nicParameters := network.Interface{
		Location: &az.azconfig.Location,
		Tags:     tags,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				network.InterfaceIPConfiguration{
					Name: to.StringPtr("ip1"),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						Subnet: &network.Subnet{
							ID: to.StringPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers"+
								"/Microsoft.Network/virtualnetworks/%s/subnets/%s",
								az.azconfig.SubscriptionID,
								az.azconfig.ResourceGroup,
								az.azconfig.Network,
								az.azconfig.Subnet)),
						},
						PrivateIPAllocationMethod: network.Dynamic,
					},
				},
			},
		},
	}
	nic, err := az.netClient.createOrUpdate(az.ctx, az.azconfig.ResourceGroup, name+"-nic", nicParameters)
	if err != nil {
		return nil, wrapAzureError(err)
	}

	instanceVhd := fmt.Sprintf("https://%s.blob.%s/%s/%s-os.vhd",
		az.azconfig.StorageAccount,
		az.azureEnv.StorageEndpointSuffix,
		az.azconfig.BlobContainer,
		name)

	customData := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`#!/bin/sh
echo '%s-%s' > /home/crunch/node-token`, name, newTags["node-token"])))

	vmParameters := compute.VirtualMachine{
		Location: &az.azconfig.Location,
		Tags:     tags,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.VirtualMachineSizeTypes(instanceType.ProviderType),
			},
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					OsType:       compute.Linux,
					Name:         to.StringPtr(name + "-os"),
					CreateOption: compute.FromImage,
					Image: &compute.VirtualHardDisk{
						URI: to.StringPtr(string(imageID)),
					},
					Vhd: &compute.VirtualHardDisk{
						URI: &instanceVhd,
					},
				},
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					compute.NetworkInterfaceReference{
						ID: nic.ID,
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
			OsProfile: &compute.OSProfile{
				ComputerName:  &name,
				AdminUsername: to.StringPtr("crunch"),
				LinuxConfiguration: &compute.LinuxConfiguration{
					DisablePasswordAuthentication: to.BoolPtr(true),
					SSH: &compute.SSHConfiguration{
						PublicKeys: &[]compute.SSHPublicKey{
							compute.SSHPublicKey{
								Path:    to.StringPtr("/home/crunch/.ssh/authorized_keys"),
								KeyData: to.StringPtr(string(ssh.MarshalAuthorizedKey(publicKey))),
							},
						},
					},
				},
				CustomData: &customData,
			},
		},
	}

	vm, err := az.vmClient.createOrUpdate(az.ctx, az.azconfig.ResourceGroup, name, vmParameters)
	if err != nil {
		return nil, wrapAzureError(err)
	}

	return &azureInstance{
		provider: az,
		nic:      nic,
		vm:       vm,
	}, nil
}

func (az *azureInstanceSet) Instances(cloud.InstanceTags) ([]cloud.Instance, error) {
	az.stopWg.Add(1)
	defer az.stopWg.Done()

	interfaces, err := az.manageNics()
	if err != nil {
		return nil, err
	}

	result, err := az.vmClient.listComplete(az.ctx, az.azconfig.ResourceGroup)
	if err != nil {
		return nil, wrapAzureError(err)
	}

	instances := make([]cloud.Instance, 0)

	for ; result.NotDone(); err = result.Next() {
		if err != nil {
			return nil, wrapAzureError(err)
		}
		if strings.HasPrefix(*result.Value().Name, az.namePrefix) {
			instances = append(instances, &azureInstance{
				provider: az,
				vm:       result.Value(),
				nic:      interfaces[*(*result.Value().NetworkProfile.NetworkInterfaces)[0].ID]})
		}
	}
	return instances, nil
}

// ManageNics returns a list of Azure network interface resources.
// Also performs garbage collection of NICs which have "namePrefix", are
// not associated with a virtual machine and have a "create-at" time
// more than DeleteDanglingResourcesAfter (to prevent racing and
// deleting newly created NICs) in the past are deleted.
func (az *azureInstanceSet) manageNics() (map[string]network.Interface, error) {
	az.stopWg.Add(1)
	defer az.stopWg.Done()

	result, err := az.netClient.listComplete(az.ctx, az.azconfig.ResourceGroup)
	if err != nil {
		return nil, wrapAzureError(err)
	}

	interfaces := make(map[string]network.Interface)

	timestamp := time.Now()
	for ; result.NotDone(); err = result.Next() {
		if err != nil {
			az.logger.WithError(err).Warnf("Error listing nics")
			return interfaces, nil
		}
		if strings.HasPrefix(*result.Value().Name, az.namePrefix) {
			if result.Value().VirtualMachine != nil {
				interfaces[*result.Value().ID] = result.Value()
			} else {
				if result.Value().Tags["created-at"] != nil {
					createdAt, err := time.Parse(time.RFC3339Nano, *result.Value().Tags["created-at"])
					if err == nil {
						if timestamp.Sub(createdAt).Seconds() > az.azconfig.DeleteDanglingResourcesAfter.Duration().Seconds() {
							az.logger.Printf("Will delete %v because it is older than %v s", *result.Value().Name, az.azconfig.DeleteDanglingResourcesAfter)
							az.deleteNIC <- *result.Value().Name
						}
					}
				}
			}
		}
	}
	return interfaces, nil
}

// ManageBlobs garbage collects blobs (VM disk images) in the
// configured storage account container.  It will delete blobs which
// have "namePrefix", are "available" (which means they are not
// leased to a VM) and haven't been modified for
// DeleteDanglingResourcesAfter seconds.
func (az *azureInstanceSet) manageBlobs() {
	result, err := az.storageAcctClient.ListKeys(az.ctx, az.azconfig.ResourceGroup, az.azconfig.StorageAccount)
	if err != nil {
		az.logger.WithError(err).Warn("Couldn't get account keys")
		return
	}

	key1 := *(*result.Keys)[0].Value
	client, err := storage.NewBasicClientOnSovereignCloud(az.azconfig.StorageAccount, key1, az.azureEnv)
	if err != nil {
		az.logger.WithError(err).Warn("Couldn't make client")
		return
	}

	blobsvc := client.GetBlobService()
	blobcont := blobsvc.GetContainerReference(az.azconfig.BlobContainer)

	page := storage.ListBlobsParameters{Prefix: az.namePrefix}
	timestamp := time.Now()

	for {
		response, err := blobcont.ListBlobs(page)
		if err != nil {
			az.logger.WithError(err).Warn("Error listing blobs")
			return
		}
		for _, b := range response.Blobs {
			age := timestamp.Sub(time.Time(b.Properties.LastModified))
			if b.Properties.BlobType == storage.BlobTypePage &&
				b.Properties.LeaseState == "available" &&
				b.Properties.LeaseStatus == "unlocked" &&
				age.Seconds() > az.azconfig.DeleteDanglingResourcesAfter.Duration().Seconds() {

				az.logger.Printf("Blob %v is unlocked and not modified for %v seconds, will delete", b.Name, age.Seconds())
				az.deleteBlob <- b
			}
		}
		if response.NextMarker != "" {
			page.Marker = response.NextMarker
		} else {
			break
		}
	}
}

func (az *azureInstanceSet) Stop() {
	az.stopFunc()
	az.stopWg.Wait()
	close(az.deleteNIC)
	close(az.deleteBlob)
}

type azureInstance struct {
	provider *azureInstanceSet
	nic      network.Interface
	vm       compute.VirtualMachine
}

func (ai *azureInstance) ID() cloud.InstanceID {
	return cloud.InstanceID(*ai.vm.ID)
}

func (ai *azureInstance) String() string {
	return *ai.vm.Name
}

func (ai *azureInstance) ProviderType() string {
	return string(ai.vm.VirtualMachineProperties.HardwareProfile.VMSize)
}

func (ai *azureInstance) SetTags(newTags cloud.InstanceTags) error {
	ai.provider.stopWg.Add(1)
	defer ai.provider.stopWg.Done()

	tags := make(map[string]*string)

	for k, v := range ai.vm.Tags {
		if !strings.HasPrefix(k, "dispatch-") {
			tags[k] = v
		}
	}
	for k, v := range newTags {
		newstr := v
		tags["dispatch-"+k] = &newstr
	}

	vmParameters := compute.VirtualMachine{
		Location: &ai.provider.azconfig.Location,
		Tags:     tags,
	}
	vm, err := ai.provider.vmClient.createOrUpdate(ai.provider.ctx, ai.provider.azconfig.ResourceGroup, *ai.vm.Name, vmParameters)
	if err != nil {
		return wrapAzureError(err)
	}
	ai.vm = vm

	return nil
}

func (ai *azureInstance) Tags() cloud.InstanceTags {
	tags := make(map[string]string)

	for k, v := range ai.vm.Tags {
		if strings.HasPrefix(k, "dispatch-") {
			tags[k[9:]] = *v
		}
	}

	return tags
}

func (ai *azureInstance) Destroy() error {
	ai.provider.stopWg.Add(1)
	defer ai.provider.stopWg.Done()

	_, err := ai.provider.vmClient.delete(ai.provider.ctx, ai.provider.azconfig.ResourceGroup, *ai.vm.Name)
	return wrapAzureError(err)
}

func (ai *azureInstance) Address() string {
	return *(*ai.nic.IPConfigurations)[0].PrivateIPAddress
}

func (ai *azureInstance) VerifyHostKey(receivedKey ssh.PublicKey, client *ssh.Client) error {
	ai.provider.stopWg.Add(1)
	defer ai.provider.stopWg.Done()

	remoteFingerprint := ssh.FingerprintSHA256(receivedKey)

	tags := ai.Tags()

	tg := tags["ssh-pubkey-fingerprint"]
	if tg != "" {
		if remoteFingerprint == tg {
			return nil
		}
		return fmt.Errorf("Key fingerprint did not match, expected %q got %q", tg, remoteFingerprint)
	}

	nodetokenTag := tags["node-token"]
	if nodetokenTag == "" {
		return fmt.Errorf("Missing node token tag")
	}

	sess, err := client.NewSession()
	if err != nil {
		return err
	}

	nodetokenbytes, err := sess.Output("cat /home/crunch/node-token")
	if err != nil {
		return err
	}

	nodetoken := strings.TrimSpace(string(nodetokenbytes))

	expectedToken := fmt.Sprintf("%s-%s", *ai.vm.Name, nodetokenTag)

	if strings.TrimSpace(nodetoken) != expectedToken {
		return fmt.Errorf("Node token did not match, expected %q got %q", expectedToken, nodetoken)
	}

	sess, err = client.NewSession()
	if err != nil {
		return err
	}

	keyfingerprintbytes, err := sess.Output("ssh-keygen -E sha256 -l -f /etc/ssh/ssh_host_rsa_key.pub")
	if err != nil {
		return err
	}

	sp := strings.Split(string(keyfingerprintbytes), " ")

	if remoteFingerprint != sp[1] {
		return fmt.Errorf("Key fingerprint did not match, expected %q got %q", sp[1], remoteFingerprint)
	}

	tags["ssh-pubkey-fingerprint"] = sp[1]
	delete(tags, "node-token")
	ai.SetTags(tags)
	return nil
}
