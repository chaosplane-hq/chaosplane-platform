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

output "master_username" {
  value = "chaosplane_admin"
}

output "master_password" {
  value     = random_password.master.result
  sensitive = true
}
