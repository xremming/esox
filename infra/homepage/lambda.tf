resource "aws_cloudwatch_log_group" "homepage" {
  name              = "/aws/lambda/abborre-${var.env}-homepage"
  retention_in_days = 14
}

locals {
  base_url_suffix = var.env == "prod" ? "" : "-${var.env}"
}

resource "aws_lambda_function" "homepage" {
  filename         = "../abborre-lambda.zip"
  source_code_hash = filebase64sha256("../abborre-lambda.zip")

  function_name = "abborre-${var.env}-homepage"
  role          = aws_iam_role.homepage.arn
  handler       = "abborre"
  runtime       = "go1.x"
  timeout       = 5
  memory_size   = 512
  publish       = false

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.homepage.name
      BASE_URL   = "https://abborre${local.base_url_suffix}.remming.org"

      // TODO: get SECRETS, OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET from AWS Secrets Manager
      SECRETS             = "unsafe-abba1234"
      OAUTH_CLIENT_ID     = ""
      OAUTH_CLIENT_SECRET = ""

      // TODO: remove once discord login is in place
      ADMIN_PASSWORD = "maximilian94"
    }
  }

  depends_on = [aws_cloudwatch_log_group.homepage]
  lifecycle {
    ignore_changes = [filename, source_code_hash]
  }
}

resource "aws_lambda_function_url" "homepage" {
  function_name      = aws_lambda_function.homepage.function_name
  authorization_type = "NONE"
  invoke_mode        = "RESPONSE_STREAM"
}
