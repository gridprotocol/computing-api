#!/bin/bash
kubectl delete deploy nginx-example --now
kubectl delete svc svc-nginx-example