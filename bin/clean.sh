#!/bin/bash

kubectl delete deploy hello-world --now
kubectl delete svc svc-hello-world

kubectl delete deploy nginx-example --now
kubectl delete svc svc-nginx-example

kubectl delete deploy ubuntu --now
kubectl delete svc svc-ubuntu