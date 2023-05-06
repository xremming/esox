variable "env" {
  type     = string
  nullable = false
}

output "lambda_arn" {
  value = aws_lambda_function.homepage.arn
}

output "function_url" {
  value = aws_lambda_function_url.homepage.function_url
}
