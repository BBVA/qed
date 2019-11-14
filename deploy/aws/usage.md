
# Deploy QED Cluster and Prometheus+Grafana on AWS

## Requirements:
* Python, Pip
* AWS Cli (pip install awscli)
* Terraform =+ 0.12 (go get github.com/hashicorp/terraform)
* Terraform Inventory =+ 0.9 (go get github.com/adammck/terraform-inventory)

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

## Testing purposes: create an isolated workspace
```
$ terraform workspace new <workspace_name>
$ terraform select <workspace_name>
```

## Deploy QED cluster with agents, storage, workload and monitoring
```
$ terraform apply -auto-approve 
```
## Deploy QED cluster, workload and monitoring
```
$ terraform apply -target=null_resource.qed-base
```
The AWS Public IP will generated as output at the end.

## Destroy
```
$ terraform destroy -auto-approve 
```
