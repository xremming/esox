data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "homepage" {
  statement {
    sid    = "DynamoDB"
    effect = "Allow"

    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:ConditionCheckItem",
      "dynamodb:DeleteItem",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:UpdateItem",
    ]

    resources = [aws_dynamodb_table.homepage.arn]
  }

  statement {
    sid    = "Logs"
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["${aws_cloudwatch_log_group.homepage.arn}:*"]
  }
}

resource "aws_iam_role" "homepage" {
  name               = "abborre-${var.env}-homepage"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_policy" "homepage" {
  name   = "abborre-${var.env}-homepage"
  policy = data.aws_iam_policy_document.homepage.json
}

resource "aws_iam_role_policy_attachment" "homepage" {
  role       = aws_iam_role.homepage.name
  policy_arn = aws_iam_policy.homepage.arn
}
