terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.47"
    }

    helm = {
      source = "hashicorp/helm"
    }
  }
}

module "karpenter" {
  source  = "terraform-aws-modules/eks/aws//modules/karpenter"
  version = "21.23.0"

  cluster_name                    = var.cluster_name
  enable_spot_termination         = true
  create_pod_identity_association = true
  create_node_iam_role            = true
  node_iam_role_name              = var.name

  node_iam_role_additional_policies = {
    AmazonSSMManagedInstanceCore = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  }

  tags = var.tags
}

resource "helm_release" "karpenter" {
  name             = var.name
  repository       = "oci://public.ecr.aws/karpenter"
  chart            = "karpenter"
  version          = "1.12.1"
  namespace        = "karpenter"
  create_namespace = true

  set = [
    {
      name  = "settings.clusterName"
      value = var.cluster_name
    },
    {
      name  = "settings.clusterEndpoint"
      value = var.cluster_endpoint
    }
  ]

  values = [
    yamlencode({
      tolerations = [
        {
          key      = "CriticalAddonsOnly"
          operator = "Exists"
        }
      ]

      nodeSelector = {
        nodepool = "system"
      }
    })
  ]
}
