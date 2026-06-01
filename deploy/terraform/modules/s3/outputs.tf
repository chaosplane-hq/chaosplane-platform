output "bucket_arns" {
  value = { for key, bucket in aws_s3_bucket.this : key => bucket.arn }
}

output "bucket_ids" {
  value = { for key, bucket in aws_s3_bucket.this : key => bucket.id }
}
