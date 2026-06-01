output "db_instance_endpoint" {
  value = module.db.db_instance_endpoint
}

output "db_instance_arn" {
  value = module.db.db_instance_arn
}

output "proxy_endpoint" {
  value = aws_db_proxy.this.endpoint
}

output "proxy_id" {
  value = aws_db_proxy.this.id
}

output "db_security_group_id" {
  value = aws_security_group.db.id
}

output "master_secret_arn" {
  value = module.db.db_instance_master_user_secret_arn
}
