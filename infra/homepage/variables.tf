variable "env" {
  type     = string
  nullable = false
}

output "function_url" {
  value = aws_lambda_function_url.homepage.function_url
}
