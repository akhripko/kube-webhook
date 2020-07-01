#!/usr/bin/env bash

NAMESPACE=${1}
if [ "${NAMESPACE}" == "" ]; then
  NAMESPACE="default"
fi
echo NAMESPACE=$NAMESPACE

HOST=${2}
if [ "${HOST}" == "" ]; then
  HOST="k8s-acl-sv"
fi
echo HOST=$HOST

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"

if [ ! -f rootCA.key ]; then
openssl genrsa -out rootCA.key 4096
fi

if [ ! -f rootCA.crt ]; then
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.crt \
 -subj "/C=US/ST=New Jersey/L=Princeton /O=Dow Jones/OU=PIB/CN=*.$NAMESPACE.svc"
fi

if [ ! -f webhook.key ]; then
openssl genrsa -out webhook.key 4096
fi

if [ ! -f webhook.csr ]; then
openssl req -new -key webhook.key -out webhook.csr \
 -subj "/C=US/ST=New Jersey/L=Princeton /O=Dow Jones/OU=PIB/CN=$HOST.$NAMESPACE.svc"
fi

if [ ! -f webhook.crt ]; then
openssl x509 -req -in webhook.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out webhook.crt -days 1024 -sha256
fi

cert=$(cat rootCA.crt | base64 | tr -d '\n')
webhookKey=$(cat webhook.key | base64 | tr -d '\n')
webhookCrt=$(cat webhook.crt | base64 | tr -d '\n')

certFileName=.helm/cert_${NAMESPACE}_${HOST}.yaml
if [ ! -f certFileName ]; then
  echo "CERT_HOST: ${HOST}.${NAMESPACE}.svc" | cat > $certFileName
  echo "webhookCABundle: ${cert}" >> $certFileName
  echo "serviceCert: ${webhookCrt}" >> $certFileName
  echo "serviceKey: ${webhookKey}" >> $certFileName
else
  sed -i "" "s|\(webhookCABundle: \).*|\1$cert|g" $certFileName
  sed -i "" "s|\(serviceCert: \).*|\1$webhookCrt|g" $certFileName
  sed -i "" "s|\(serviceKey: \).*|\1$webhookKey|g" $certFileName
fi

rm rootCA.crt
rm rootCA.key
rm rootCA.srl
rm webhook.crt
rm webhook.csr
rm webhook.key