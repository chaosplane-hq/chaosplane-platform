data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  azs                 = slice(data.aws_availability_zones.available.names, 0, 3)
  flow_log_bucket_arn = var.flow_log_bucket_arn != null ? var.flow_log_bucket_arn : aws_s3_bucket.flow_logs[0].arn
}

resource "aws_s3_bucket" "flow_logs" {
  count = var.flow_log_bucket_arn == null ? 1 : 0

  bucket        = "${var.name}-vpc-flow-logs"
  force_destroy = false
  tags          = var.tags
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "6.6.1"

  name = var.name
  cidr = "10.0.0.0/16"

  azs             = local.azs
  public_subnets  = ["10.0.0.0/20", "10.0.16.0/20", "10.0.32.0/20"]
  private_subnets = ["10.0.48.0/20", "10.0.64.0/20", "10.0.80.0/20"]

  enable_nat_gateway     = true
  single_nat_gateway     = true
  one_nat_gateway_per_az = false

  enable_dns_hostnames = true
  enable_dns_support   = true

  enable_flow_log           = true
  flow_log_destination_type = "s3"
  flow_log_destination_arn  = local.flow_log_bucket_arn

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
    "karpenter.sh/discovery"          = var.name
  }

  tags = var.tags
}
