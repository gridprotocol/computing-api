#!/bin/bash

#kubectl delete deployment --all -l app/clean=label-for-clean

kubectl delete deploy --all
kubectl delete svc --all