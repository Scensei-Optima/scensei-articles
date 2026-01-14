resource "aws_efs_file_system" "efs" {
  creation_token = var.efs_creation_token
  encrypted      = true
}

resource "aws_security_group" "efs" {
  name        = "efs-sg"
  description = "Allow NFS traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 2049
    to_port     = 2049
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr_block]
  }

  # EFS might need to communicate with internal AWS logging or health-check services
  # Adjust as needed
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_efs_mount_target" "efs_mt" {
  for_each        = zipmap(module.vpc.azs, module.vpc.private_subnets)
  file_system_id  = aws_efs_file_system.efs.id
  subnet_id       = each.value
  security_groups = [aws_security_group.efs.id]
}
