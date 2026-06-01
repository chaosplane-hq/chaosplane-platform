terraform {
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      version               = "~> 6.47"
      configuration_aliases = [aws.us_east_1]
    }
  }
}

resource "aws_route53_zone" "this" {
  name = var.domain
  tags = var.tags
}

resource "aws_acm_certificate" "this" {
  domain_name               = "*.${var.domain}"
  subject_alternative_names = [var.domain]
  validation_method         = "DNS"
  tags                      = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate" "us_east_1" {
  provider = aws.us_east_1

  domain_name               = "*.${var.domain}"
  subject_alternative_names = [var.domain]
  validation_method         = "DNS"
  tags                      = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "validation" {
  for_each = {
    for option in aws_acm_certificate.this.domain_validation_options : option.domain_name => {
      name   = option.resource_record_name
      record = option.resource_record_value
      type   = option.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.this.zone_id
}

resource "aws_route53_record" "validation_us_east_1" {
  for_each = {
    for option in aws_acm_certificate.us_east_1.domain_validation_options : option.domain_name => {
      name   = option.resource_record_name
      record = option.resource_record_value
      type   = option.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.this.zone_id
}

resource "aws_acm_certificate_validation" "this" {
  certificate_arn         = aws_acm_certificate.this.arn
  validation_record_fqdns = [for record in aws_route53_record.validation : record.fqdn]
}

resource "aws_acm_certificate_validation" "us_east_1" {
  provider = aws.us_east_1

  certificate_arn         = aws_acm_certificate.us_east_1.arn
  validation_record_fqdns = [for record in aws_route53_record.validation_us_east_1 : record.fqdn]
}
