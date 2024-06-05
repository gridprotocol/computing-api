#!/bin/bash
kubectl delete deploy nginx --now
kubectl delete svc svc-nginx