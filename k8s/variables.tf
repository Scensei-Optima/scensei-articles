variable "efs_filesystem_id" {
  type = string
  description = "ID of the EFS filesystem. You can get it with `terraform output` in the `prepare` catalog"
}

variable "should_create_test_pods" {
  type = bool
  description = "Whether to create pods and PVCs to test if storage attachment was successful."
  default = false
}
