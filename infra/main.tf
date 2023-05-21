locals {
  homepage = {
    dev = {
      aliases             = ["dev.${var.domain_name}"],
      detailed_monitoring = false,
    },
    test = {
      aliases             = ["test.${var.domain_name}"],
      detailed_monitoring = false,
    },
    prod = {
      aliases             = ["www.${var.domain_name}", "${var.domain_name}"],
      detailed_monitoring = true,
    },
  }
}

data "aws_route53_zone" "default" {
  name = var.domain_name
}

module "homepage" {
  for_each = local.homepage

  source = "./homepage"

  env                 = each.key
  aliases             = each.value.aliases
  zone_id             = data.aws_route53_zone.default.zone_id
  cert_arn            = aws_acm_certificate.cert.arn
  detailed_monitoring = each.value.detailed_monitoring
}
