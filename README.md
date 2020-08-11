ChubaoMonitorOperator

Prerequisites:
```
go version v1.13+.
docker version 17.03+.
kubectl version v1.11.3+.
kustomize v3.1.0+
operator-SDK v0.19.0
Access to a Kubernetes v1.11.3+ cluster
```


Build and run ChubaoMonitorOperator in master node locally outside the cluster:
```
git clone https://github.com/Hats-Wang/ChubaoMonitorOperator.git
cd ChubaoMonitorOperator
make generate && make manifests && make install && make run ENABLE_WEBHOOKS=false
```

Edit file ./config/samples/cache_v1alpha1_chubaomonitor.yaml to customize your own ChubaoMonitor instance:


Create your own ChubaoMonitor instance:
```
kubectl create -f ./config/samples/cache_v1alpha1_chubaomonitor.yaml
```

CRD definition: api/v1alpha1/chubaomonitor_types.go

Business logic code are located in folder ./controllers


Reference： https://v0-19-x.sdk.operatorframework.io/docs/golang/quickstart/
