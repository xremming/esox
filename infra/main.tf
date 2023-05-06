locals {
  homepage = {
    dev  = null,
    test = null,
    prod = null,
  }
}

module "homepage" {
  for_each = local.homepage

  source = "./homepage"
  env    = each.key
}
