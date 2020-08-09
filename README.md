ChubaoMonitorOperator

CRD定义位于./api/v1alpha1/chubaomonitor_types.go

业务逻辑处理程序位于./controllers文件夹中

yaml文件位于config/samples/cache_v1alpha1_chubaomonitor.yaml

本地快速运行： make run ENABLE_WEBHOOKS=false

创建CR： kubectl create -f config/samples/cache_v1alpha1_chubaomonitor.yaml

Reference： https://sdk.operatorframework.io/docs/golang/quickstart/
