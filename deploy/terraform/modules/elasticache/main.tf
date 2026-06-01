resource "aws_elasticache_subnet_group" "this" {
  name       = var.name
  subnet_ids = var.subnet_ids

  tags = var.tags
}

resource "aws_security_group" "this" {
  name   = "${var.name}-elasticache"
  vpc_id = var.vpc_id

  tags = var.tags
}

resource "aws_vpc_security_group_ingress_rule" "redis" {
  for_each = toset(var.allowed_security_group_ids)

  security_group_id            = aws_security_group.this.id
  referenced_security_group_id = each.value
  from_port                    = 6379
  to_port                      = 6379
  ip_protocol                  = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.this.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = var.name
  description                = var.name
  engine                     = "valkey"
  engine_version             = "8.1"
  node_type                  = var.node_type
  num_cache_clusters         = 1
  port                       = 6379
  at_rest_encryption_enabled = true
  kms_key_id                 = var.kms_key_arn
  transit_encryption_enabled = true
  subnet_group_name          = aws_elasticache_subnet_group.this.name
  security_group_ids         = [aws_security_group.this.id]

  tags = var.tags
}
