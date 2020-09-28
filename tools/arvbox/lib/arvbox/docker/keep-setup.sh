#!/bin/bash
# Copyright (C) The Arvados Authors. All rights reserved.
#
# SPDX-License-Identifier: AGPL-3.0

exec 2>&1
sleep 2
set -eux -o pipefail

. /usr/local/lib/arvbox/common.sh
. /usr/local/lib/arvbox/go-setup.sh

flock /var/lib/gopath/gopath.lock go install "git.arvados.org/arvados.git/services/keepstore"
install $GOPATH/bin/keepstore /usr/local/bin

if test "$1" = "--only-deps" ; then
    exit
fi

mkdir -p $ARVADOS_CONTAINER_PATH/$1

export ARVADOS_API_HOST=$localip:${services[controller-ssl]}
export ARVADOS_API_HOST_INSECURE=1
export ARVADOS_API_TOKEN=$(cat $ARVADOS_CONTAINER_PATH/superuser_token)

set +e
read -rd $'\000' keepservice <<EOF
{
 "service_host":"localhost",
 "service_port":$2,
 "service_ssl_flag":false,
 "service_type":"disk"
}
EOF
set -e

if test -s $ARVADOS_CONTAINER_PATH/$1-uuid ; then
    keep_uuid=$(cat $ARVADOS_CONTAINER_PATH/$1-uuid)
    arv keep_service update --uuid $keep_uuid --keep-service "$keepservice"
else
    UUID=$(arv --format=uuid keep_service create --keep-service "$keepservice")
    echo $UUID > $ARVADOS_CONTAINER_PATH/$1-uuid
fi

management_token=$(cat $ARVADOS_CONTAINER_PATH/management_token)

set +e
sv hup /var/lib/arvbox/service/keepproxy

cat >$ARVADOS_CONTAINER_PATH/$1.yml <<EOF
Listen: "localhost:$2"
BlobSigningKeyFile: $ARVADOS_CONTAINER_PATH/blob_signing_key
SystemAuthTokenFile: $ARVADOS_CONTAINER_PATH/superuser_token
ManagementToken: $management_token
MaxBuffers: 20
EOF

exec /usr/local/bin/keepstore -config=$ARVADOS_CONTAINER_PATH/$1.yml
