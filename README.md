# node-instance-decorator
K8s operator that automatically names EC2 node instances for easier identification as K8s worker nodes

## Description
The instance decorator solves the problem of identifying which EC2 instances belong to EKS K8s worker nodes by automatically updating their Name label. These can then be easily viewed/searched using the AWS EC2 web console or CLI.

Nodes are named using the pattern:
    eks-{clusterName}-{nodeGroupName}-ip-{vpcIpAddress}

Cluster deployment is a Helm chart.

Node instance decorator is intended to be run as a singleton within a given cluster. There is no advantage to running multiple instances.

## Prerequisites
You will need:
    - An AWS EKS cluster with at least one node group, and the ARN associated with this cluster.
    - An AWS ECR repository in which to deploy the operator image, and the URI associated with this repository. 

### Installing into the cluster

1. Create an IAM role and associated policy that grants permission for the relevant EC2 operations.
   
```sh
script-runner scripts\nodeInstanceDecorator-prepare-config -p "cluster.arn:{CLUSTER_ARN}"
```

Note the ARN of the role that is created.


1. Build and push your image to ECR.
	
```sh
make docker-build docker-push REPO_URI={REPOSITORY_URI}
```
	
1. Deploy the controller to the cluster.

```sh
make deploy REPO_URI={REPOSITORY_URI} CLUSTER_ARN={CLUSTER_ARN} ROLE_ARN={ROLE_ARN}
```

Existing worker nodes should be processed and their names updated automatically in the EC2 instance list.


### Uninstallation
Remove the operator from the cluster using:

```sh
make undeploy
```

### How it works
This project uses the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)


### Debugging 
To debug the Helm chart and inspect the intermediate yaml file that is created run:

```sh
make helm-debug
```

Output will be generated as debug.yaml

