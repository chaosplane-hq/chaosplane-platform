output "distribution_ids" {
  value = {
    for key, distribution in aws_cloudfront_distribution.site : key => distribution.id
  }
}

output "distribution_domain_names" {
  value = {
    for key, distribution in aws_cloudfront_distribution.site : key => distribution.domain_name
  }
}

output "bucket_ids" {
  value = {
    for key, bucket in aws_s3_bucket.site : key => bucket.id
  }
}
