# node-instance-decorator
K8s operator that automatically names EC2 node instances for easier identification as K8s worker nodes.

<br/>

## Description
The instance decorator solves the problem of identifying which EC2 instances belong to EKS K8s worker nodes by automatically updating their Name label. These can then be easily viewed/searched using the AWS EC2 web console or CLI.

Nodes are named using the pattern:
    {ClusterName}-eks-{NodeGroupName}-workerNode-{NodeIPAddress} ({Zone}, {OperatingSystem})

<br/>

## Operator scope

node-instance-decorator does *not* scope its activities to the namespace in which it is installed. You do not need to (and should not) install copies into different namespaces within a single cluster.

A single installation will manage node naming across an entire cluster. 

<br/>

## Cluster installation

node-instance-decorator consists of two components:
- An image containing the operator
- A Helm chart containing the K8s deployment

Both must be installed to use the operator within a cluster.

### Prerequisites
You will need:
- An AWS EKS cluster with at least one node group, and the ARN associated with this cluster.
- An AWS ECR repository into which to deploy the operator image, and the URI associated with this repository. 
- Local installations of golang, kubectl, aws-cli and helm. On Windows, these should be installed within WSL.
- Local installation of script-runner, if intending to use scripted IAM role/policy creation.

    **NOTE:** The .kubeconfig associated with the WSL kubectl is *NOT* the same as the one used in Windows. 
Verify cluster access within WSL using `kubectl config get-contexts` and, if  necessary, add the required context using e.g. `aws --region {aws.region} eks update-kubeconfig --name {cluster.name}`. 

### Installation procedure

1. Create an IAM role and associated policy that grants permission for the relevant EC2 operations. Note the ARN of the role that is created.

   - Script runner can automate this task:
   
    ```
        script-runner scripts\nodeInstanceDecorator-prepare-config -p "cluster.arn:{CLUSTER_ARN}"
    ```

   - To do this manually, follow the instructions at https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html using the trust policy template `scripts\_resources\nodeInstanceDecorator-iam-role-trust-policy.template`.

    <br/>
    
2. Build and push the operator image to ECR.
	
    ```sh
        make docker-build docker-push REPO_URI={REPOSITORY_URI}
    ```

    **NOTE:** On Windows, run this command within WSL.
	
    <br/>
    
3. Deploy the operator to the cluster.

    ```sh
        make deploy REPO_URI={REPOSITORY_URI} CLUSTER_ARN={CLUSTER_ARN} ROLE_ARN={ROLE_ARN}
    ```
    
    **NOTE:** On Windows, run this command within WSL.

    Existing worker nodes should be processed and their corresponding EC2 instance names updated automatically. You can view these names using e.g. the AWS EC2 web  console or CLI.

<br/>

## Using in Kubernetes

The operator is fully automated - no action is required once installation is complete.

Review the EC2 worker node instances to confirm that their names are being automatically configured.

<br/>

## Configuration

You can modify how worker nodes are named by configuring the `config.nameTemplate` key in `values.yaml`.

A name template is a string containing one or more substitution parameters delimited by curly braces, e.g. `{Zone}`.

Valid substitution parameters are: {Zone}, {ClusterName}, {NodeGroupName}, {NodeIPAddress}, {HostName}, {OperatingSystem}, {Architecture}

<br/>

## Uninstallation
Remove the operator from the cluster using:

```sh
    make undeploy CLUSTER_ARN={CLUSTER_ARN}
```

<br/>

## How it works
This project uses the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It was built from a kubebuilder project and subsequently modified to use Helm.

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

<br/>

## Debugging 
To debug the Helm chart and inspect the intermediate yaml file that is created run:

```sh
make helm-debug
```

Output will be generated as `debug.yaml`

