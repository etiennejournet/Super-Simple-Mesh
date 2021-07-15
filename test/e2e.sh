#!/bin/bash

curl -sfL https://get.k3s.io | sh -
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl /usr/local/bin

export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"

kubectl create ns super-simple-mesh
helm install cert-manager jetstack/cert-manager --set installCRDs=true -n super-simple-mesh
kubectl apply -f deploy/manifest

go test -v
