#!/usr/bin/env bash

echo "::group::Get nodes"
kubectl get nodes
echo "::endgroup::"

echo "::group::Describe nodes"
kubectl describe nodes
echo "::endgroup::"

echo "::group::Get all events"
kubectl get events -A
echo "::endgroup::"

echo "::group::Get KServe pods"
kubectl get pods -n kserve -o wide
kubectl get pods -n kserve -o yaml
echo "::endgroup::"

echo "::group::KServe log"
kubectl logs deployment/kserve-controller-manager -n kserve -c manager
echo "::endgroup::"

echo "::group::Get Knative pods"
kubectl get pods -n knative-serving -o wide
kubectl get pods -n knative-serving -o yaml
echo "::endgroup::"

echo "::group::Knative log"
kubectl logs deploy/controller -n knative-serving
echo "::endgroup::"

echo "::group::Get deployments in caraml namespace"
kubectl get deployment -n caraml
echo "::endgroup::"

echo "::group::Describe pods in caraml namespace"
kubectl describe pods -n caraml
echo "::endgroup::"

echo "::group::Merlin log"
kubectl logs deploy/merlin -n caraml -c merlin
echo "::endgroup::"

echo "::group::Get Inference Service in merlin-e2e namespace"
kubectl get isvc -n merlin-e2e -o wide
echo "::endgroup::"

echo "::group::Describe Inference Service in merlin-e2e namespace"
kubectl describe isvc -n merlin-e2e
echo "::endgroup::"

echo "::group::Get deployment in merlin-e2e namespace"
kubectl get deployment -n merlin-e2e
echo "::endgroup::"

echo "::group::Describe deployment in merlin-e2e namespace"
kubectl describe deployment -n merlin-e2e
echo "::endgroup::"

echo "::group::Get pods in merlin-e2e namespace"
kubectl get pod -n merlin-e2e -o yaml
echo "::endgroup::"

echo "::group::Describe pods in merlin-e2e namespace"
kubectl describe pod -n merlin-e2e
echo "::endgroup::"

echo "::group::Get events in merlin-e2e namespace"
kubectl get events -n merlin-e2e
echo "::endgroup::"

echo "::group::Models log"
kubectl logs -l component=predictor -c kserve-container -n merlin-e2e
echo "::endgroup::"

echo "::group::Queue proxy log"
kubectl logs -l component=predictor -c queue-proxy -n merlin-e2e
echo "::endgroup::"

echo "::group::Transformers log"
kubectl logs -l component=transformer -c kserve-container -n merlin-e2e
echo "::endgroup::"
