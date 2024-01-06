data "aws_route53_zone" "default" {
  for_each = toset(local.domains)

  name = each.value
}

locals {
  domains = flatten([var.canonical_domain, var.additional_domains])

  homepage = {
    dev = {
      aliases = [for domain in local.domains : {
        name    = "dev.${domain}",
        zone_id = data.aws_route53_zone.default[domain].zone_id,
      }]
      detailed_monitoring = false,
    },
    test = {
      aliases = [for domain in local.domains : {
        name    = "test.${domain}",
        zone_id = data.aws_route53_zone.default[domain].zone_id,
      }],
      detailed_monitoring = false,
    },
    prod = {
      aliases = flatten([for domain in local.domains : [{
        name    = "${domain}",
        zone_id = data.aws_route53_zone.default[domain].zone_id,
        }, {
        name    = "www.${domain}",
        zone_id = data.aws_route53_zone.default[domain].zone_id,
      }]]),
      detailed_monitoring = true,
    },
  }
}

module "homepage" {
  for_each = local.homepage

  source = "./homepage"

  env                 = each.key
  aliases             = each.value.aliases
  cert_arn            = aws_acm_certificate.cert.arn
  detailed_monitoring = each.value.detailed_monitoring
}
