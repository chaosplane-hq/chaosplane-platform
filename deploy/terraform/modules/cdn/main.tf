resource "aws_s3_bucket" "site" {
  for_each = var.sites

  bucket = "${var.name}-${each.key}"

  tags = var.tags
}

resource "aws_s3_bucket_public_access_block" "site" {
  for_each = aws_s3_bucket.site

  bucket = each.value.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_cloudfront_origin_access_control" "site" {
  for_each = var.sites

  name                              = "${var.name}-${each.key}"
  description                       = "${var.name}-${each.key}"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "site" {
  for_each = var.sites

  enabled             = true
  default_root_object = "index.html"
  aliases             = [each.value.domain]

  origin {
    domain_name              = aws_s3_bucket.site[each.key].bucket_regional_domain_name
    origin_id                = aws_s3_bucket.site[each.key].id
    origin_access_control_id = aws_cloudfront_origin_access_control.site[each.key].id
  }

  default_cache_behavior {
    target_origin_id       = aws_s3_bucket.site[each.key].id
    viewer_protocol_policy = "redirect-to-https"
    compress               = true
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = var.certificate_arn_us_east_1
    ssl_support_method  = "sni-only"
  }

  tags = var.tags
}

data "aws_iam_policy_document" "site" {
  for_each = var.sites

  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.site[each.key].arn}/*"]

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.site[each.key].arn]
    }
  }
}

resource "aws_s3_bucket_policy" "site" {
  for_each = var.sites

  bucket = aws_s3_bucket.site[each.key].id
  policy = data.aws_iam_policy_document.site[each.key].json
}
