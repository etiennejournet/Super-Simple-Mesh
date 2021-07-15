#!/bin/bash
export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"

curl -sfL https://get.k3s.io | sh - --write-kubeconfig-mode 644
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

kubectl create ns super-simple-mesh
kubectl apply -f deploy/manifest
helm install cert-manager jetstack/cert-manager --set installCRDs=true -n super-simple-mesh

go test -v
