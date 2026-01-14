resource "kubernetes_namespace_v1" "scensei_namespace" {
  metadata {
    name = "scensei"
  }
}

data "terraform_remote_state" "eks_state" {
  backend = "local"
  config = {
    path = "../terraform.tfstate"
  }
}

resource "kubernetes_storage_class" "efs_sc" {
  metadata {
    name = "efs-sc"
  }

  storage_provisioner = "efs.csi.aws.com"

  parameters = {
    provisioningMode = "efs-ap"
    fileSystemId     = data.terraform_remote_state.eks_state.outputs.efs_id
    directoryPerms   = "700"
    gidRangeStart    = "1000"
    gidRangeEnd      = "2000"
    basePath         = "/dynamic_provisioning"
  }

  reclaim_policy      = "Retain"
  volume_binding_mode = "Immediate"
}

resource "kubernetes_storage_class" "ebs_sc" {
  metadata {
    name = "ebs-sc"
  }

  storage_provisioner = "ebs.csi.aws.com"

  parameters = {
    type      = "gp3"
    fsType    = "ext4"
    encrypted = "true"
    # Optional: kmsKeyId = "alias/aws/ebs"
  }

  volume_binding_mode = "WaitForFirstConsumer"
}

resource "kubernetes_persistent_volume_claim" "scenarios" {
  metadata {
    name      = "efs-scenarios"
    namespace = kubernetes_namespace_v1.scensei_namespace.metadata[0].name
  }

  spec {
    access_modes       = ["ReadWriteMany"]
    storage_class_name = "efs-sc"
    resources {
      requests = {
        storage = "5Gi"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "logs" {
  metadata {
    name      = "efs-logs"
    namespace = kubernetes_namespace_v1.scensei_namespace.metadata[0].name
  }

  spec {
    access_modes       = ["ReadWriteMany"]
    storage_class_name = "efs-sc"
    resources {
      requests = {
        storage = "5Gi"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "trajectories" {
  metadata {
    name      = "efs-trajectories"
    namespace = kubernetes_namespace_v1.scensei_namespace.metadata[0].name
  }

  spec {
    access_modes       = ["ReadWriteMany"]
    storage_class_name = "efs-sc"
    resources {
      requests = {
        storage = "5Gi"
      }
    }
  }
}
