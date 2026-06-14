output "zone_id" {
  value = aws_route53_zone.this.zone_id
}

output "zone_name_servers" {
  value = aws_route53_zone.this.name_servers
}

output "certificate_arn" {
  value = aws_acm_certificate.this.arn
}

output "certificate_arn_us_east_1" {
  value = aws_acm_certificate.us_east_1.arn
}
