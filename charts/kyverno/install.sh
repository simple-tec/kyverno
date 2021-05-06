#!/bin/bash
# for helm2
# helm install --namespace=logging --name=log-slave  ./
# for helm3
helm install --namespace=kyverno kyverno  ./
