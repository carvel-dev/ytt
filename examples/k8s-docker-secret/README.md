This example was created to show how to configure `kubernetes.io/dockerconfigjson` k8s secret based on data values (https://kubernetes.slack.com/archives/CH8KCCKA5/p1560891111037300).

```bash
export CFG_docker__server=https://docker.io
export CFG_docker__username=dkalinin
export CFG_docker__password=secret
ytt -f config.yml -f schema.yml --data-values-env CFG
```
