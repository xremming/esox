resource "aws_cloudwatch_log_group" "homepage" {
  name              = "/aws/lambda/abborre-${var.env}-homepage"
  retention_in_days = 14
}

resource "aws_lambda_function" "homepage" {
  filename         = "../abborre-lambda.zip"
  source_code_hash = filebase64sha256("../abborre-lambda.zip")

  function_name = "abborre-${var.env}-homepage"
  role          = aws_iam_role.homepage.arn
  handler       = "abborre-lambda"
  runtime       = "go1.x"
  timeout       = 5
  memory_size   = 512
  publish       = false

  depends_on = [aws_cloudwatch_log_group.homepage]

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.homepage.name
    }
  }
}

resource "aws_lambda_function_url" "homepage" {
  function_name      = aws_lambda_function.homepage.function_name
  authorization_type = "NONE"
  invoke_mode        = "RESPONSE_STREAM"
}
