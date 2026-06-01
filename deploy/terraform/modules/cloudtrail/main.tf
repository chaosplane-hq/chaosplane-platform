data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_s3_bucket" "trail" {
  bucket = "${var.name}-cloudtrail"
  tags   = var.tags
}

resource "aws_s3_bucket_policy" "trail" {
  bucket = aws_s3_bucket.trail.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSCloudTrailAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.trail.arn
      },
      {
        Sid    = "AWSCloudTrailWrite"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.trail.arn}/AWSLogs/${data.aws_caller_identity.current.account_id}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}

resource "aws_cloudtrail" "this" {
  name                          = var.name
  s3_bucket_name                = aws_s3_bucket.trail.id
  kms_key_id                    = var.kms_key_arn
  is_multi_region_trail         = true
  include_global_service_events = true
  enable_log_file_validation    = true
  tags                          = var.tags

  depends_on = [aws_s3_bucket_policy.trail]
}
