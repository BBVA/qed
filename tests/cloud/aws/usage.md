
# Cloud benchmarking on AWS

## Requirements:
* Python, Pip
* AWS Cli (pip install awscli)
* Terraform (go get github.com/hashicorp/terraform)

## Execute the benchmark

Modify the value of the TF_FLAVOUR variable in run-benchmarks script. ie: TF_FLAVOUR="t2.2xlarge c5.2xlarge"

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
