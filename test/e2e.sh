#!/bin/bash
export KUBECONFIG="kubeconfig"

curl -sfL https://get.k3s.io | sh -
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl /usr/local/bin

kubectl create ns super-simple-mesh
kubectl apply -f deploy/manifest
helm install cert-manager jetstack/cert-manager --set installCRDs=true -n super-simple-mesh

go test -v
