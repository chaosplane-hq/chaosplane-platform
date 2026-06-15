data "aws_lb" "this" {
  name = var.alb_name
}

resource "aws_cloudfront_distribution" "this" {
  for_each = var.sites

  enabled = true
  comment = "${var.name}-${each.key}"

  aliases = each.value.domain != null ? [each.value.domain] : []

  origin {
    domain_name = data.aws_lb.this.dns_name
    origin_id   = "alb"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }

    custom_header {
      name  = "X-Origin-Verify"
      value = var.origin_verify_secret
    }

    custom_header {
      name  = "X-Site"
      value = each.key
    }
  }

  default_cache_behavior {
    target_origin_id       = "alb"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"]
    cached_methods         = ["GET", "HEAD"]

    forwarded_values {
      query_string = true
      headers      = ["*"]

      cookies {
        forward = "all"
      }
    }

    min_ttl     = 0
    default_ttl = 0
    max_ttl     = 0
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = each.value.domain == null
    acm_certificate_arn            = each.value.domain != null ? var.acm_certificate_arn : null
    ssl_support_method             = each.value.domain != null ? "sni-only" : null
    minimum_protocol_version       = each.value.domain != null ? "TLSv1.2_2021" : null
  }
}
