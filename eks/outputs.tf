output "efs_id" {
  description = "ID of the EFS - needed to create the storage class that will underlie PVCs"
  value       = aws_efs_file_system.efs.id
}
