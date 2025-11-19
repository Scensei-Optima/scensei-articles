# Provision EKS cluster and mount EFS and EBS persistent volume to nodes

## Prerequisites

Run the command

```shell
terraform init
```

in both `prepare` and `deploy` catalog.

Also make sure you have `aws cli` installed.

---

## Create cluster, node group and EFS

The following steps should result in creation of an EKS cluster with mounted EFS and EBS volumes.

### Create VPC

Navigate to `prepare` catalog and apply the terraform configuration for the VPC*.

```shell
terraform apply -target module.vpc
```

Type yes + RETURN KEY and wait until the command finishes.

You can confirm in the web console that the VPC `optima-eks` was created in region `eu-north-1`.

*mounting EFS VPC's requires terraform to get the subnets IDs on the plan-time, not create-time


### Create the cluster and storage

Apply remaining terraform resources

```shell
terraform apply
```

Type yes + RETURN KEY and wait until the command finishes.

You can confirm in the console that the EKS cluster and EFS disk were created.

### Get the EFS filesystem ID

This will be needed in the next step.

```shell
export TF_VAR_efs_filesystem_id = "$(terraform output -raw efs_filesystem_id)"
```

If you perform this action in this way, you can use the same terminal session to create the K8s resources
described below.

---

## Create K8s resources

In this step you will provision the `scensei` namespace, storage classes, PVCs and the postgres operator release
to the cluster.

If you want to create test pods that will allow you to make sure that the storage attachment was successful,
run this command beforehand:

```shell
export TF_VAR_should_create_test_pods=true
```

Navigate to the `deploy/` catalog and apply the terraform configuration.

```shell
terraform apply
```

After the command finished you can check for the created resources using `kubectl`.


### [Optional] Verify test pods and PVCs

Check if PVC has status `Bound`.

```shell
kubectl get pvc
```

You should see something like:

```terminaloutput
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
ebs-claim   Bound    pvc-c299f337-9c25-4ef6-a777-c5b5527d8760   5Gi        RWO            ebs-sc         <unset>                 4s
efs-claim   Bound    pvc-3f617818-1513-4007-9848-84499dcd988e   5Gi        RWX            efs-sc         <unset>                 5s
```

---
!! **IMPORTANT** !!

**The EFS PVC created here is only to test if you can interact with EFS from K8s pods.**

**In actual OPTIMA scenario you will create a PVC with the `ReadOnlyMany` access mode.**

---

Exec on the pod and write a test file to `/mnt/efs` (mounting point in the test pod):

```shell
kubectl exec --stdin --tty efs-test -- /bin/bash

# on the pod
echo "TEST" > /mnt/efs/test
```

Then when you delete the pod and create it again:

```shell
kubectl exec --stdin --tty efs-test -- /bin/bash

# on the pod
cat /mnt/efs/test
```

You should see:

```terminaloutput
bash-5.2# cat /mnt/efs/test
TEST
```

If this succeeded then your EKS cluster has EFS storage correctly mounted on the EC2 nodes.


## [Optional] Deleting the cluster

If you don't need the cluster anymore, just run

```shell
terraform destroy -auto-approve
```

in both `prepare` and `deploy` catalogs and wait until the command completes.
