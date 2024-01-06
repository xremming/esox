resource "aws_route53_record" "a" {
  count = length(var.aliases)

  zone_id = var.aliases[count.index].zone_id
  name    = var.aliases[count.index].name
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.default.domain_name
    zone_id                = aws_cloudfront_distribution.default.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "aaaa" {
  count = length(var.aliases)

  zone_id = var.aliases[count.index].zone_id
  name    = var.aliases[count.index].name
  type    = "AAAA"

  alias {
    name                   = aws_cloudfront_distribution.default.domain_name
    zone_id                = aws_cloudfront_distribution.default.hosted_zone_id
    evaluate_target_health = false
  }
}
