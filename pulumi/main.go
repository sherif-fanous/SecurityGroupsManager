package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cloudWatchLogsPolicy, err := iam.NewPolicy(ctx, "SecurityGroupsManagerLambdaFunctionCloudWatchLogsPolicy", &iam.PolicyArgs{
			Path: pulumi.String("/"),
			Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Action": [
							"logs:CreateLogGroup",
							"logs:CreateLogStream",
							"logs:PutLogEvents"
						],
						"Resource": "*",
						"Effect": "Allow"
					}
				]
			}`),
		})
		if err != nil {
			return err
		}

		ec2Policy, err := iam.NewPolicy(ctx, "SecurityGroupsManagerLambdaFunctionEC2Policy", &iam.PolicyArgs{
			Path: pulumi.String("/"),
			Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Action": [
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
						"Resource": "*",
						"Effect": "Allow"
					}
				]
			}`),
		})
		if err != nil {
			return err
		}

		role, err := iam.NewRole(ctx, "SecurityGroupsManagerLambdaFunctionRole", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
				  {
					"Effect": "Allow",
					"Principal": {
					  "Service": "lambda.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
				  }
				]
			  }`),
			ManagedPolicyArns: pulumi.StringArray{
				cloudWatchLogsPolicy.Arn,
				ec2Policy.Arn,
			},
			Path: pulumi.String("/"),
		})
		if err != nil {
			return err
		}

		lambdaFunction, err := lambda.NewFunction(ctx, "SecurityGroupsManager", &lambda.FunctionArgs{
			Code: pulumi.NewFileArchive("ManagedSecurityGroups.zip"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"CONFIGURATION": pulumi.String(config.Require(ctx, "CONFIGURATION")),
					"DEBUG":         pulumi.String(config.Require(ctx, "DEBUG")),
				},
			},
			Handler: pulumi.String("main"),
			Role:    role.Arn,
			Runtime: pulumi.String("go1.x"),
			Timeout: pulumi.Int(30),
		})
		if err != nil {
			return err
		}

		scheduledEventRule, err := cloudwatch.NewEventRule(ctx, "SecurityGroupsManagerLambdaFunctionScheduledEventRule", &cloudwatch.EventRuleArgs{
			ScheduleExpression: pulumi.String(config.Require(ctx, "SCHEDULE_EXPRESSION")),
		})
		if err != nil {
			return err
		}

		_, err = lambda.NewPermission(ctx, "SecurityGroupsManagerLambdaFunctionScheduledEventRulePermission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  lambdaFunction.Name,
			Principal: pulumi.String("events.amazonaws.com"),
			SourceArn: scheduledEventRule.Arn,
		})
		if err != nil {
			return err
		}

		_, err = cloudwatch.NewEventTarget(ctx, "SecurityGroupsManager", &cloudwatch.EventTargetArgs{
			Arn:  lambdaFunction.Arn,
			Rule: scheduledEventRule.Name,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
