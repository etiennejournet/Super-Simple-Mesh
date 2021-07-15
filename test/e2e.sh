#!/bin/bash
export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"

curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE="644" sh -s -
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

## Create Namespace
kubectl create ns super-simple-mesh

## Setup Cert-Manager using Helm
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager --set installCRDs=true -n super-simple-mesh

## Setup SSM
kubectl apply -f ../deploy/manifest

## Setup ClusterIssuer
kubectl apply -f ./manifest/clusterissuer.yml

go test -v
