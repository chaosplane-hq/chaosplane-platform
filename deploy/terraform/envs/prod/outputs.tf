output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "cluster_name" {
  value = module.eks.cluster_name
}

output "ecr_repository_urls" {
  value = module.ecr.repository_urls
}

output "rds_proxy_endpoint" {
  value = module.rds.proxy_endpoint
}

output "zone_name_servers" {
  value = module.dns.zone_name_servers
}

output "waf_acl_arn" {
  value = module.waf.web_acl_arn
}

output "github_actions_role_arn" {
  value = module.github_oidc.role_arn
}
