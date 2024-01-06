resource "aws_acm_certificate" "cert" {
  provider = aws.us-east-1

  domain_name       = var.canonical_domain
  validation_method = "DNS"
  subject_alternative_names = concat(
    ["*.${var.canonical_domain}"],
    flatten([for domain in var.additional_domains : ["${domain}", "*.${domain}"]]),
  )

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "cert" {
  for_each = {
    for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  zone_id         = data.aws_route53_zone.default[trimprefix(each.key, "*.")].zone_id
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
}

resource "aws_acm_certificate_validation" "cert" {
  provider = aws.us-east-1

  depends_on              = [aws_route53_record.cert]
  certificate_arn         = aws_acm_certificate.cert.arn
  validation_record_fqdns = [for record in aws_route53_record.cert : record.fqdn]
}
