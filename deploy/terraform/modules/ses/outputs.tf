output "ses_domain_identity_arn" {
  value = aws_ses_domain_identity.this.arn
}

output "smtp_endpoint" {
  value = "email-smtp.${data.aws_region.current.region}.amazonaws.com"
}
