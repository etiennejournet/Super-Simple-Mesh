#!/bin/bash

export KUBECONFIG=~/SCHNEIDER/azure-infra/azure/terraform/generated/dev-kubeconfig
kubectl apply -f test.yml

kubectl wait --for=condition=Ready pod/test
kubectl wait --for=condition=Ready pod/test2

echo "test -- apt update"
kubectl exec -t test -- apt update > /dev/null 2>&1 || echo "fail"
echo "test -- curl test2"
kubectl exec -t test -- curl test2 2> /dev/null
echo "test -- curl test2.default.svc.cluster.local"
kubectl exec -t test -- curl test2.default.svc.cluster.local 2> /dev/null

echo "test2 -- apt update"
kubectl exec -t test2 -- apt update > /dev/null 2>&1 || echo "fail"
echo "test2 -- curl test"
kubectl exec -t test2 -- curl test:8000 2> /dev/null
echo "test2 -- curl test.default.svc.cluster.local:8000"
kubectl exec -t test2 -- curl test.default.svc.cluster.local:8000 2> /dev/null

kubectl delete -f test.yml
