data "tls_certificate" "github_actions_oidc_provider" {
  url = "https://token.actions.githubusercontent.com/.well-known/openid-configuration"
}

# Needs to be re-ran periodically because the certificate will expire periodically
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
  client_id_list = [
    "https://github.com/supercell",
    "sts.amazonaws.com",
  ]

  thumbprint_list = [
    data.tls_certificate.github_actions_oidc_provider.certificates[0].sha1_fingerprint,
  ]
}

data "aws_iam_policy_document" "github_actions" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:xremming/abborre:*"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.github.arn]
      type        = "Federated"
    }
  }
}

data "aws_iam_policy_document" "github_actions_policy" {
  statement {
    sid    = "UpdateFunctionCode"
    effect = "Allow"

    actions   = ["lambda:UpdateFunctionCode", "lambda:GetFunctionConfiguration"]
    resources = [for homepage in module.homepage : homepage.lambda_arn]
  }
}

resource "aws_iam_role" "github_actions" {
  name = "GitHub-Actions"

  assume_role_policy = data.aws_iam_policy_document.github_actions.json
}

resource "aws_iam_role_policy" "github_actions" {
  role        = aws_iam_role.github_actions.name
  name_prefix = "GitHub-Actions"
  policy      = data.aws_iam_policy_document.github_actions_policy.json
}

output "role_name" {
  value = aws_iam_role.github_actions.name
}
