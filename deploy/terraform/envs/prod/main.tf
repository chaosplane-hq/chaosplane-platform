terraform {
  required_version = ">= 1.15.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.47"
    }
  }

  backend "s3" {
    bucket       = "chaosplane-prod-terraform-state"
    key          = "prod/terraform.tfstate"
    region       = "ap-northeast-2"
    profile      = "chaosplane"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region  = var.region
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = "chaosplane-hq"
    }
  }
}

provider "aws" {
  alias   = "us_east_1"
  region  = "us-east-1"
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = "chaosplane-hq"
    }
  }
}

variable "domain" {
  type    = string
  default = "chaosplane.com"
}

variable "alert_emails" {
  type    = list(string)
  default = []
}

locals {
  name = "${var.project}-${var.environment}"
}

module "kms" {
  source = "../../modules/kms"

  name = local.name
  tags = {}
}

module "vpc" {
  source = "../../modules/vpc"

  name = local.name
  tags = {}
}

module "vpc_endpoints" {
  source = "../../modules/vpc-endpoints"

  name                    = local.name
  vpc_id                  = module.vpc.vpc_id
  vpc_cidr                = module.vpc.vpc_cidr_block
  private_subnet_ids      = module.vpc.private_subnet_ids
  private_route_table_ids = module.vpc.private_route_table_ids
  tags                    = {}

  depends_on = [module.vpc]
}

module "ecr" {
  source = "../../modules/ecr"

  repositories = ["chaosplane/api", "chaosplane/web"]
  kms_key_arn  = module.kms.key_arn
  tags         = {}

  depends_on = [module.kms]
}

module "eks" {
  source = "../../modules/eks"

  name        = local.name
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.private_subnet_ids
  kms_key_arn = module.kms.key_arn
  tags        = {}

  depends_on = [module.vpc, module.kms]
}

module "karpenter" {
  source = "../../modules/karpenter"

  name                   = local.name
  cluster_name           = module.eks.cluster_name
  cluster_endpoint       = module.eks.cluster_endpoint
  node_security_group_id = module.eks.node_security_group_id
  tags                   = {}

  depends_on = [module.eks]
}

module "rds" {
  source = "../../modules/rds"

  name        = local.name
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.private_subnet_ids
  vpc_cidr    = module.vpc.vpc_cidr_block
  kms_key_arn = module.kms.key_arn
  tags        = {}

  depends_on = [module.vpc, module.kms]
}

module "elasticache" {
  source = "../../modules/elasticache"

  name        = local.name
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.private_subnet_ids
  vpc_cidr    = module.vpc.vpc_cidr_block
  kms_key_arn = module.kms.key_arn
  tags        = {}

  depends_on = [module.vpc]
}

module "s3" {
  source = "../../modules/s3"

  name = local.name
  buckets = {
    reports = { versioning = true }
    logs    = { versioning = false }
    ai      = { versioning = false }
  }
  kms_key_arn = module.kms.key_arn
  tags        = {}

  depends_on = [module.kms]
}

module "dns" {
  source = "../../modules/dns"

  providers = {
    aws.us_east_1 = aws.us_east_1
  }

  domain = var.domain
  tags   = {}
}

module "ses" {
  source = "../../modules/ses"

  domain  = var.domain
  zone_id = module.dns.zone_id
  tags    = {}

  depends_on = [module.dns]
}

module "secrets" {
  source = "../../modules/secrets"

  name = local.name
  secrets = {
    "db-url"      = "postgres://${module.rds.master_username}:${module.rds.master_password}@${module.rds.proxy_endpoint}:5432/chaosplane?sslmode=require"
    "redis-url"   = "rediss://${module.elasticache.endpoint}:${module.elasticache.port}"
    "jwt-secret"  = ""
    "csrf-secret" = ""
  }
  kms_key_arn = module.kms.key_arn
  tags        = {}

  depends_on = [module.kms, module.elasticache, module.rds]
}

module "pod_identity" {
  source = "../../modules/pod-identity"

  name                  = local.name
  cluster_name          = module.eks.cluster_name
  rds_proxy_resource_id = module.rds.proxy_id
  s3_bucket_arns        = values(module.s3.bucket_arns)
  hosted_zone_arns      = ["arn:aws:route53:::hostedzone/${module.dns.zone_id}"]
  tags                  = {}

  depends_on = [module.eks, module.rds, module.s3, module.dns]
}

module "waf" {
  source = "../../modules/waf"

  name = local.name
  tags = {}
}

module "guardduty" {
  source = "../../modules/guardduty"

  name         = local.name
  alert_emails = var.alert_emails
  tags         = {}
}

module "cloudtrail" {
  source = "../../modules/cloudtrail"

  name        = local.name
  kms_key_arn = module.kms.key_arn
  tags        = {}

  depends_on = [module.kms]
}
