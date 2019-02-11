
# Deploy QED Cluster and Prometheus+Grafana on AWS

## Requirements:
* Python, Pip
* AWS Cli (pip install awscli)
* Terraform (go get github.com/hashicorp/terraform)

## Init 

```
$ export GO111MODULE=on
$ terraform init -backend-config "$aws-profile=name"

```

## Deploy
```
$ terraform apply -auto-approve 
```
The AWS Public IP will generated as output at the end.

## Destroy
```
$ terraform destroy -auto-approve 
```
