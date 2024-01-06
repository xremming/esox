variable "canonical_domain" {
  type     = string
  nullable = false
  default  = "abborre.fi"
}

variable "additional_domains" {
  type     = list(string)
  nullable = false
  default  = ["abborrefreediving.fi", "abborre.net"]
}
