resource "aws_cloudfront_function" "forward_host" {
  name    = "forward-host"
  runtime = "cloudfront-js-2.0"
  comment = "Forwards the original Host header as X-Original-Host."
  publish = true
  code    = file("${path.module}/forward_host.js")
}
