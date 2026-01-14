variable "efs_creation_token" {
  type        = string
  description = "Creation token for the EFS"
  default     = "optima-efs"
}

variable "should_update_kubeconfig" {
  type        = bool
  description = "Whether to automatically update the kubeconfig after EKS creation"
  default     = true
}

variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster"
  default     = "optima"
}

variable "kubernetes_version" {
  type        = string
  description = "K8s version for the EKS cluster"
  default     = "1.34"
}
