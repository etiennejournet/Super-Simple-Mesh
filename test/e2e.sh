#!/bin/bash
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

k3s kubectl create ns super-simple-mesh
k3s kubectl apply -f deploy/manifest
helm install cert-manager jetstack/cert-manager --set installCRDs=true -n super-simple-mesh

go test -v
