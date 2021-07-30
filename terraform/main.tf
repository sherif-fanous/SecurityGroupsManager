resource "random_id" "suffix" {
  byte_length = 4
}

resource "aws_iam_policy" "security_groups_manager_cloud_watch_logs_policy" {
  name = "SecurityGroupsManagerLambdaFunctionCloudWatchLogsPolicy-${random_id.suffix.id}"
  path = "/"

  policy = jsonencode({
    Version : "2012-10-17",
    Statement : [
      {
        Action : [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource : "*",
        Effect : "Allow"
      }
    ]
  })
}

resource "aws_iam_policy" "security_groups_manager_ec2_policy" {
  name = "SecurityGroupsManagerLambdaFunctionEC2Policy-${random_id.suffix.id}"
  path = "/"

  policy = jsonencode({
    Version : "2012-10-17",
    Statement : [
      {
        Action : [
          "ec2:AuthorizeSecurityGroupEgress",
          "ec2:AuthorizeSecurityGroupIngress",
          "ec2:CreateTags",
          "ec2:DeleteTags",
          "ec2:DescribeRegions",
          "ec2:DescribeSecurityGroups",
          "ec2:RevokeSecurityGroupEgress",
          "ec2:RevokeSecurityGroupIngress",
          "ec2:UpdateSecurityGroupRuleDescriptionsEgress",
          "ec2:UpdateSecurityGroupRuleDescriptionsIngress"
        ],
        Resource : "*",
        Effect : "Allow"
      }
    ]
  })
}

resource "aws_iam_role" "security_groups_manager_execution_role" {
  managed_policy_arns = [aws_iam_policy.security_groups_manager_cloud_watch_logs_policy.arn, aws_iam_policy.security_groups_manager_ec2_policy.arn]
  name                = "SecurityGroupsManagerLambdaFunctionRole-${random_id.suffix.id}"
  path                = "/"

  assume_role_policy = jsonencode({
    Version : "2012-10-17",
    Statement : [
      {
        Effect : "Allow",
        Principal : {
          Service : "lambda.amazonaws.com"
        },
        Action : "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_lambda_function" "security_groups_manager" {
  filename      = "ManagedSecurityGroups.zip"
  function_name = "SecurityGroupsManager-${random_id.suffix.id}"
  handler       = "main"
  role          = aws_iam_role.security_groups_manager_execution_role.arn
  runtime       = "go1.x"
  timeout       = 30

  environment {
    variables = {
      CONFIGURATION = var.configuration
      DEBUG         = var.debug
    }
  }
}

resource "aws_cloudwatch_event_rule" "security_groups_manager_scheduled_event_rule" {
  name                = "SecurityGroupsManagerLambdaFunctionScheduledEventRule-${random_id.suffix.id}"
  schedule_expression = var.schedule_expression
}

resource "aws_cloudwatch_event_target" "security_groups_manager_scheduled_event_rule_target" {
  arn  = aws_lambda_function.security_groups_manager.arn
  rule = aws_cloudwatch_event_rule.security_groups_manager_scheduled_event_rule.name
}

resource "aws_lambda_permission" "security_groups_manager_scheduled_event_rule_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.security_groups_manager.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.security_groups_manager_scheduled_event_rule.arn
}
