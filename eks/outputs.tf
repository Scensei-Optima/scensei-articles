output "efs_filesystem_id" {
  value = aws_efs_file_system.efs.id
  description = "ID of the EFS filesystem."
}
