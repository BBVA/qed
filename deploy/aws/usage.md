
# Deploy QED Cluster and Prometheus+Grafana on AWS

## Requirements:
* Python, Pip
* AWS Cli (pip install awscli)
* Terraform (go get github.com/hashicorp/terraform)

## Init 

```
$ export GO111MODULE=on
$ terraform init -backend-config "profile=${your_aws_profile}"

```

## Bandaid
If terraform misbehaves, give it a gentle nudge like this:
```
$ terraform init -backend-config "profile=${your_aws_profile} -reconfigure"
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
