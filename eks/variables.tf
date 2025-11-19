variable "should_update_kubeconfig" {
  type = bool
  description = "Whether to automatically update the kubeconfig after EKS creation"
  default = true
}
