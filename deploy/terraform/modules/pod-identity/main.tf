data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

module "lb_controller" {
  source = "terraform-aws-modules/eks-pod-identity/aws"

  name                            = "${var.name}-lbc"
  attach_aws_lb_controller_policy = true

  associations = {
    this = {
      cluster_name    = var.cluster_name
      namespace       = "kube-system"
      service_account = "aws-load-balancer-controller"
    }
  }

  tags = var.tags
}

module "external_secrets" {
  source = "terraform-aws-modules/eks-pod-identity/aws"

  name                           = "${var.name}-external-secrets"
  attach_external_secrets_policy = true

  associations = {
    this = {
      cluster_name    = var.cluster_name
      namespace       = "kube-system"
      service_account = "external-secrets"
    }
  }

  tags = var.tags
}

module "cert_manager" {
  source = "terraform-aws-modules/eks-pod-identity/aws"

  name                       = "${var.name}-cert-manager"
  attach_cert_manager_policy = true

  associations = {
    this = {
      cluster_name    = var.cluster_name
      namespace       = "cert-manager"
      service_account = "cert-manager"
    }
  }

  tags = var.tags
}

module "fluent_bit" {
  source = "terraform-aws-modules/eks-pod-identity/aws"

  name                 = "${var.name}-fluent-bit"
  attach_custom_policy = true

  policy_statements = [
    {
      effect = "Allow"
      actions = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams"
      ]
      resources = ["*"]
    }
  ]

  associations = {
    this = {
      cluster_name    = var.cluster_name
      namespace       = "monitoring"
      service_account = "fluent-bit"
    }
  }

  tags = var.tags
}

module "chaosplane_api" {
  source = "terraform-aws-modules/eks-pod-identity/aws"

  name                 = "${var.name}-chaosplane-api"
  attach_custom_policy = true
  policy_statements = [
    {
      effect = "Allow"
      actions = [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ]
      resources = ["*"]
    },
    {
      effect    = "Allow"
      actions   = ["s3:ListBucket"]
      resources = var.s3_bucket_arns
    },
    {
      effect = "Allow"
      actions = [
        "s3:PutObject",
        "s3:GetObject"
      ]
      resources = [for arn in var.s3_bucket_arns : "${arn}/*"]
    },
    {
      effect    = "Allow"
      actions   = ["rds-db:connect"]
      resources = ["arn:aws:rds-db:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:dbuser:${var.rds_proxy_resource_id}/*"]
    }
  ]

  associations = {
    this = {
      cluster_name    = var.cluster_name
      namespace       = "chaosplane-system"
      service_account = "chaosplane-api"
    }
  }

  tags = var.tags
}

module "external_dns" {
  source = "terraform-aws-modules/eks-pod-identity/aws"

  name                          = "${var.name}-external-dns"
  attach_external_dns_policy    = true
  external_dns_hosted_zone_arns = var.hosted_zone_arns

  associations = {
    this = {
      cluster_name    = var.cluster_name
      namespace       = "kube-system"
      service_account = "external-dns"
    }
  }

  tags = var.tags
}
