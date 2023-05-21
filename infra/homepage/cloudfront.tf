data "aws_region" "current" {}

resource "aws_cloudfront_cache_policy" "default" {
  name = "abborre-homepage-${var.env}"

  min_ttl     = 0
  default_ttl = 0
  max_ttl     = 24 * 60 * 60

  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["session", "flash"]
      }
    }

    headers_config {
      header_behavior = "none"
    }

    query_strings_config {
      query_string_behavior = "all"
    }
  }
}

resource "aws_cloudfront_distribution" "default" {
  enabled             = true
  wait_for_deployment = false

  comment = "abborre homepage - ${var.env}"

  is_ipv6_enabled = true
  http_version    = "http3"
  price_class     = "PriceClass_All"

  aliases = var.aliases

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = var.cert_arn
    ssl_support_method  = "sni-only"
  }

  origin {
    origin_id   = "Default"
    domain_name = "${aws_lambda_function_url.homepage.url_id}.lambda-url.${data.aws_region.current.name}.on.aws"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id = "Default"

    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"]
    cached_methods         = ["GET", "HEAD", "OPTIONS"]

    cache_policy_id            = aws_cloudfront_cache_policy.default.id
    origin_request_policy_id   = null
    response_headers_policy_id = null
  }
}

moved {
  from = aws_cloudfront_distribution.s3_distribution
  to   = aws_cloudfront_distribution.default
}

moved {
  from = aws_cloudfront_monitoring_subscription.example
  to   = aws_cloudfront_monitoring_subscription.default
}

resource "aws_cloudfront_monitoring_subscription" "default" {
  count = var.detailed_monitoring ? 1 : 0

  distribution_id = aws_cloudfront_distribution.default.id

  monitoring_subscription {
    realtime_metrics_subscription_config {
      realtime_metrics_subscription_status = "Enabled"
    }
  }
}
