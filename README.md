[![Build Status](https://travis-ci.com/sfanous/SecurityGroupsManager.svg?branch=master)](https://travis-ci.com/sfanous/SecurityGroupsManager)
[![Release](https://img.shields.io/github/v/release/sfanous/SecurityGroupsManager.svg?style=flat)](https://github.com/sfanous/SecurityGroupsManager/releases/latest)

# SecurityGroupsManager

## Overview

SecurityGroupsManager is an AWS serverless application. The core of the application is a Lambda Function written in Go.

## Problem Statment

The premise for SecurityGroupsManager is simple.

You provide a desired state of one or more security groups and SecurityGroupsManager monitors the security groups in question and ensures they are always kept in sync with the desired state.

Additionally, SecurityGroupsManager addresses AWS' security groups limitation where the source or destination of a security group rule requires any of the following:

1. An individual IPv4 or IPv6 address, in CIDR block notation
2. A range of IPv4 or IPv6 addresses, in CIDR block notation
3. A prefix list ID
4. Another security group

The list above excludes fully qualified domain names (FQDN). This is an issue for home users like myself who don't have a static IP address they can configure their security group rules with.

With no static IP address, the simplest solution commonly resorted to is to create wide open security group rules using 0.0.0.0/0 as the source. This allows the whole world to send traffic to the destination port but goes against the principle of least privilege.

A dynamic DNS service takes your dynamic IP address and makes it act as though it is static by pointing a static hostname to it. A dynamic DNS update client running on your PC or router is configured to periodically check for changes to your IP address. If your IP address changes, the dynamic DNS update client updates your dynamic DNS hostname with the current IP address.

## Solution

You've determined the desired state of your security groups and setup a dynamic DNS hostname. You've made sure the dynamic DNS hostname is always updated to point to your currently assigned dynamic IP address. Now all you need is a solution that monitors your security groups and ensures they are always kept in sync with the desired state.

That's where SecurityGroupsManager comes in. You provide the Lambda Function with a configuration that describes the desired state of one or more security group. On each invocation the Lamda Function compares the as is state against the desired state of each configured security group and executes any required remediations.

## Architecture
![Imgur](https://i.imgur.com/E652VQq.png)

## Deployment

The easiest way to deploy SecurityGroupsManager is to use the CloudFormation Quick Create Stack Launch URL.

[![](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?templateURL=https://manage-security-groups-cloudformation-artifacts.s3.ca-central-1.amazonaws.com/template.yaml&stackName=SecurityGroupsManagerStack)

This will open the CloudFormation Quick Create Stack Console.

![Imgur](https://i.imgur.com/J7udell.png)

You'll need to make sure you're in the AWS region in which you want CloudFormation to create your resources. To switch regions, choose the region list to the right of your account information on the navigation bar.

## CloudFormation Stack Setup

- **Stack name**
  - This is the name of the CloudFormation Stack that will be created
  - There's no need to make any changes to here, but if you feel that you want to name your stack differently then Feel free to change the value from the default **SecurityGroupsManagerStack**
- **Configuration**
  - This parameter sets the initial value of the Lambda Function's CONFIGURATION environment variable. After the Lambda Function is created you can always update the value of the CONFIGURATION environment variable from the Lambda Function console
  - This is the where you enter your security groups desired state
  - The configuration is a superset of the `aws ec2 describe-security-groups` JSON output. Here's a sample configuration

    ```json
    {
      "SecurityGroups": [
        {
          "Description": "Test SG",
          "GroupId": "sg-6d9a02303c07f74e2",
          "GroupName": "Test SG",
          "IpPermissions": [
            {
              "FromPort": 22,
              "Hosts": [
                {
                  "Description": "My home IP address",
                  "FQDN": "myHome.hopto.org"
                }
              ],
              "IpProtocol": "tcp",
              "Ipv6Ranges": [],
              "PrefixListIds": [],
              "ToPort": 22,
              "UserIdGroupPairs": []
            },
            {
              "FromPort": 80,
              "Hosts": [
                {
                  "Description": "My home IP address",
                  "FQDN": "myHome.hopto.org"
                }
              ],
              "IpProtocol": "tcp",
              "IpRanges": [
                {
                  "CidrIp": "1.2.3.4/32"
                }
              ],
              "Ipv6Ranges": [],
              "PrefixListIds": [],
              "ToPort": 80,
              "UserIdGroupPairs": []
            }
          ],
          "IpPermissionsEgress": [
            {
              "IpProtocol": "-1",
              "IpRanges": [
                {
                  "CidrIp": "0.0.0.0/0"
                }
              ],
              "Ipv6Ranges": [],
              "PrefixListIds": [],
              "UserIdGroupPairs": []
            }
          ],
          "OwnerId": "467087866041",
          "VpcId": "vpc-c86ad37e"
        }
      ]
    }
    ```

    The new addition to this JSON structure is the `Hosts` array within the `IpPermissions` or `IpPermissionsEgress` objects. This is where you configure your dynamic DNS hostname using the `FQDN` attribute. You can define as many host objects within the `Hosts` array as you need.

    In this sample we're defining the following desired state

    - Ingress rule for TCP port 22 to allow traffic from the IPv4 address pointed to by `myHome.hopto.org`
    - Ingress rule for TCP port 80 to allow traffic from the IPv4 address pointed to by `myHome.hopto.org` and the static IPv4 address `1.2.3.4`

    The Lambda Function resolves the dynamic DNS hostnames defined using the `FQDN` attribute of each host within the `Hosts` array to IPv4 & IPv6 addresses in CIDR notation and merges the results with any pre-configured `CidrIp` within the `IpRanges` and `Ipv6Ranges` respectively to create a consolidated `IpRanges` and `Ipv6Ranges` arrays then proceeds to compare the desired state with the configured state. In case of a discrepancy the current remediations are determined and applied.

  - The simplest way to create this configuration is as follows
    - Execute `aws ec2 describe-security-groups` and copy the full JSON output to your favorite editor
    - Add a `Hosts` array to the security group rules you want the Lambda Function to monitor and update
    - Copy your edited JSON and paste it into the **Configuration** parameter

- **EnableDebugMode**
  - This parameter sets the initial value of the Lambda Function's DEBUG environment variable. After the Lambda Function is created you can always update the value of the DEBUG environment variable from the Lambda Function console
  
- **RateExpressionMinutes**
  - This parameter configure the rate expression of the EventBridge rule. The Lambda Function is invoked by an EventBridge rule and this parameter controls the frequency of invocations

## Sample Output

SecurityGroupsManager sends its output to CloudWatch Logs. The output is displayed in tabular form.

<span style="color:red">**The tabular form in CloudWatch Logs will look completely messed up. This is due to text wrapping.**</span>

To be able to view the tabular form as intended
1. Check the `View as text` checkbox
2. Select the whole output for a request (From the beginning of the line that starts with `START RequestId` to the end of the line that starts with `END RequestId`)
3. Copy the selected output
4. Paste the copied output into your favorite text editor and disable text wrapping

Here are some sample outputs:

- Remediations determined and applied

![svgur](https://svgshare.com/i/YxU.svg)

- No remediations determined

![svgur](https://svgshare.com/i/Yw4.svg)

- No matching security group found

![svgur](https://svgshare.com/i/YwV.svg)

## Important Notes

- If SecurityGroupsManager encounters a configued security group for which it is unable to find a matching security group in AWS then SecurityGroupsManager will report this as seen in the last sample output. SecurityGroupsManager will not create a new security group in this case.
