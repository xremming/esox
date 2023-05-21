variable "env" {
  type     = string
  nullable = false
}

variable "aliases" {
  type     = list(string)
  nullable = false

  validation {
    condition     = length(var.aliases) > 0
    error_message = "aliases must contain at least one value"
  }
}

variable "zone_id" {
  type     = string
  nullable = false
}

variable "cert_arn" {
  type     = string
  nullable = false
}

variable "detailed_monitoring" {
  type     = bool
  nullable = false
  default  = false
}

output "lambda_arn" {
  value = aws_lambda_function.homepage.arn
}
