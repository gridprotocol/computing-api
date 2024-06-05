#!/bin/bash

kubectl delete deploy hello-world --now
kubectl delete svc svc-hello-world

kubectl delete deploy nginx --now
kubectl delete svc svc-nginx

kubectl delete deploy ubuntu --now
kubectl delete svc svc-ubuntu