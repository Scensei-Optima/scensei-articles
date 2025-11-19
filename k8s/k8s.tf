resource "kubernetes_namespace" "scensei_namespace"{
  metadata {
    name = "scensei"
  }
}

resource "kubernetes_storage_class" "efs_sc" {
  metadata {
    name = "efs-sc"
  }

  storage_provisioner = "efs.csi.aws.com"

  parameters = {
    provisioningMode = "efs-ap"
    fileSystemId     = var.efs_filesystem_id
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
    type   = "gp3"
    fsType = "ext4"
  }

  volume_binding_mode = "WaitForFirstConsumer"
}

resource "kubernetes_persistent_volume_claim" "efs_pvc" {
  metadata {
    name      = "efs-claim"
    namespace = "scensei"
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

resource "helm_release" "postgres_operator" {
  name       = "postgres-operator"
  repository = "https://opensource.zalando.com/postgres-operator/charts/postgres-operator"
  chart      = "postgres-operator"
  namespace  = "postgres"

  create_namespace = true

  set = [{
    name  = "configKubernetes.enable_cross_namespace_secret"
    value = "true"
  }]
}

# OPTIONAL TEST RESOURCES TO VERIFY STORAGE ATTACHMENT

# --- EFS PVC ---
resource "kubernetes_persistent_volume_claim" "efs_claim" {
  metadata {
    name = "efs-claim"
  }

  count = var.should_create_test_pods ? 1 : 0

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

# --- EFS test Pod ---
resource "kubernetes_pod" "efs_test" {
  metadata {
    name = "efs-test"
  }

  count = var.should_create_test_pods ? 1 : 0

  spec {
    container {
      name    = "app"
      image   = "amazonlinux"
      command = ["sh", "-c", "sleep 3600"]

      volume_mount {
        name       = "efs-vol"
        mount_path = "/mnt/efs"
      }
    }

    volume {
      name = "efs-vol"

      persistent_volume_claim {
        claim_name = kubernetes_persistent_volume_claim.efs_claim[count.index].metadata[0].name
      }
    }
  }
}
