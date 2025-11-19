module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  name               = "optima"
  kubernetes_version = "1.34"

  # optional addons
  addons = {
    coredns                = {}
    kube-proxy             = {}
    vpc-cni                = { before_compute = true }
    eks-pod-identity-agent = { before_compute = true }
    aws-ebs-csi-driver     = { service_account_role_arn = module.ebs_csi_irsa.arn }
    aws-efs-csi-driver     = { service_account_role_arn = module.efs_csi_irsa.arn }
  }

  enable_cluster_creator_admin_permissions = true
  endpoint_public_access                   = true

  compute_config = {
    enabled = false
  }

  vpc_id                   = module.vpc.vpc_id
  subnet_ids               = module.vpc.private_subnets
  control_plane_subnet_ids = module.vpc.private_subnets

  # ---------------------------
  # Node group definition
  # ---------------------------
  eks_managed_node_groups = {
    default = {
      instance_types = ["t3.medium"]
      ami_type       = "BOTTLEROCKET_x86_64"
      desired_size   = 4
      min_size       = 4
      max_size       = 4

      # optionally allow SSH access
      # remote_access = {
      #   ec2_ssh_key = "my-keypair" # replace with your existing EC2 key name
      # }

      tags = {
        Name = "optima-ng"
      }
    }
  }

  tags = {
    Environment = "dev"
    Terraform   = "true"
    Project     = "optima"
  }
}

resource "null_resource" "update_kubeconfig" {
  depends_on = [module.eks]
  count = var.should_update_kubeconfig ? 1 : 0

  provisioner "local-exec" {
    command = "aws eks update-kubeconfig --region eu-north-1 --name ${module.eks.cluster_name}"
  }
}
