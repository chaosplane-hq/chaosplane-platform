data "aws_region" "current" {}

locals {
  interface_endpoints = toset([
    "ec2",
    "ecr.api",
    "ecr.dkr",
    "sts",
    "secretsmanager",
    "logs",
  ])
}

resource "aws_security_group" "this" {
  name        = "${var.name}-vpce"
  description = "${var.name}-vpce"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.name}-vpce"
  })
}

resource "aws_vpc_endpoint" "gateway" {
  vpc_id            = var.vpc_id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = var.private_route_table_ids

  tags = merge(var.tags, {
    Name = "${var.name}-s3"
  })
}

resource "aws_vpc_endpoint" "interface" {
  for_each = local.interface_endpoints

  vpc_id              = var.vpc_id
  service_name        = "com.amazonaws.${data.aws_region.current.region}.${each.key}"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = var.private_subnet_ids
  security_group_ids  = [aws_security_group.this.id]
  private_dns_enabled = true

  tags = merge(var.tags, {
    Name = "${var.name}-${replace(each.key, ".", "-")}"
  })
}
