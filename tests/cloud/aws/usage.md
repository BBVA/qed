
# Cloud benchmarking on AWS

## Requirements:
* Python, Pip
* AWS Cli (pip install awscli)
* Terraform (go get github.com/hashicorp/terraform)

## Execute the benchmark

Modify the value of the TF_FLAVOUR variable in run-benchmarks script. ie: FLAVOURS="t2.2xlarge c5.2xlarge"
Set TF_CLUSTER_SIZE variable to 3 to run the benchmar in multi-node mode. ie: TF_CLUSTER_SIZE 3

```
./run-benchmarks
```

## Check results

The results are published in the results direcotry like this:
```
results
├── c5.2xlarge-results.txt
└── t2.2xlarge-results.txt
```
