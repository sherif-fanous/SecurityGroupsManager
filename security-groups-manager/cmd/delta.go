package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type SecurityGroupDelta struct {
	AsIsSecurityGroup                    *types.SecurityGroup
	IpPermissionsToAuthorize             []types.IpPermission
	IpPermissionsToAuthorizeResult       string
	IpPermissionsToRevoke                []types.IpPermission
	IpPermissionsToRevokeResult          string
	IpPermissionsToUpdate                []types.IpPermission
	IpPermissionsToUpdateResult          string
	IpPermissionsEgressToAuthorize       []types.IpPermission
	IpPermissionsEgressToAuthorizeResult string
	IpPermissionsEgressToRevoke          []types.IpPermission
	IpPermissionsEgressToRevokeResult    string
	IpPermissionsEgressToUpdate          []types.IpPermission
	IpPermissionsEgressToUpdateResult    string
	RegionName                           string
	TagsToCreate                         []types.Tag
	TagsToCreateResult                   string
	TagsToDelete                         []types.Tag
	TagsToDeleteResult                   string
	ToBeSecurityGroup                    *types.SecurityGroup
}

func NewSecurityGroupDelta(toBeSecurityGroup *types.SecurityGroup) *SecurityGroupDelta {
	securityGroupDelta := new(SecurityGroupDelta)

	securityGroupDelta.AsIsSecurityGroup = nil
	securityGroupDelta.IpPermissionsToAuthorize = make([]types.IpPermission, 0)
	securityGroupDelta.IpPermissionsToAuthorizeResult = ""
	securityGroupDelta.IpPermissionsToRevoke = make([]types.IpPermission, 0)
	securityGroupDelta.IpPermissionsToRevokeResult = ""
	securityGroupDelta.IpPermissionsToUpdate = make([]types.IpPermission, 0)
	securityGroupDelta.IpPermissionsToUpdateResult = ""
	securityGroupDelta.IpPermissionsEgressToAuthorize = make([]types.IpPermission, 0)
	securityGroupDelta.IpPermissionsEgressToAuthorizeResult = ""
	securityGroupDelta.IpPermissionsEgressToRevoke = make([]types.IpPermission, 0)
	securityGroupDelta.IpPermissionsEgressToRevokeResult = ""
	securityGroupDelta.IpPermissionsEgressToUpdate = make([]types.IpPermission, 0)
	securityGroupDelta.IpPermissionsEgressToUpdateResult = ""
	securityGroupDelta.TagsToCreate = make([]types.Tag, 0)
	securityGroupDelta.TagsToCreateResult = ""
	securityGroupDelta.TagsToDelete = make([]types.Tag, 0)
	securityGroupDelta.TagsToDeleteResult = ""
	securityGroupDelta.ToBeSecurityGroup = toBeSecurityGroup

	return securityGroupDelta
}

func (s *SecurityGroupDelta) apply(client *ec2.Client) {
	log.Printf("Applying remediations")

	if len(s.IpPermissionsToRevoke) > 0 {
		if _, err := client.RevokeSecurityGroupIngress(context.TODO(), &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       s.AsIsSecurityGroup.GroupId,
			IpPermissions: s.IpPermissionsToRevoke,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.IpPermissionsToRevokeResult = fmt.Sprintf("Failed to revoke inbound rules: %v", err)
		} else {
			s.IpPermissionsToRevokeResult = "Succeeded to revoke inbound rules"
		}
	}

	if len(s.IpPermissionsToAuthorize) > 0 {
		if _, err := client.AuthorizeSecurityGroupIngress(context.TODO(), &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       s.ToBeSecurityGroup.GroupId,
			IpPermissions: s.IpPermissionsToAuthorize,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.IpPermissionsToAuthorizeResult = fmt.Sprintf("Failed to authorize inbound rules: %v", err)
		} else {
			s.IpPermissionsToAuthorizeResult = "Succeeded to authorize inbound rules"
		}
	}

	if len(s.IpPermissionsToUpdate) > 0 {
		if _, err := client.UpdateSecurityGroupRuleDescriptionsIngress(context.TODO(), &ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
			GroupId:       s.ToBeSecurityGroup.GroupId,
			IpPermissions: s.IpPermissionsToUpdate,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.IpPermissionsToUpdateResult = fmt.Sprintf("Failed to update inbound rules: %v", err)
		} else {
			s.IpPermissionsToUpdateResult = "Succeeded to update inbound rules"
		}
	}

	if len(s.IpPermissionsEgressToRevoke) > 0 {
		if _, err := client.RevokeSecurityGroupEgress(context.TODO(), &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       s.AsIsSecurityGroup.GroupId,
			IpPermissions: s.IpPermissionsEgressToRevoke,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.IpPermissionsEgressToRevokeResult = fmt.Sprintf("Failed to revoke outbound rules: %v", err)
		} else {
			s.IpPermissionsEgressToRevokeResult = "Succeeded to revoke outbound rules"
		}
	}

	if len(s.IpPermissionsEgressToAuthorize) > 0 {
		if _, err := client.AuthorizeSecurityGroupEgress(context.TODO(), &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       s.ToBeSecurityGroup.GroupId,
			IpPermissions: s.IpPermissionsEgressToAuthorize,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.IpPermissionsEgressToAuthorizeResult = fmt.Sprintf("Failed to authorize outbound rules: %v", err)
		} else {
			s.IpPermissionsEgressToAuthorizeResult = "Succeeded to authorize outbound rules"
		}
	}

	if len(s.IpPermissionsEgressToUpdate) > 0 {
		if _, err := client.UpdateSecurityGroupRuleDescriptionsEgress(context.TODO(), &ec2.UpdateSecurityGroupRuleDescriptionsEgressInput{
			GroupId:       s.ToBeSecurityGroup.GroupId,
			IpPermissions: s.IpPermissionsEgressToUpdate,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.IpPermissionsEgressToUpdateResult = fmt.Sprintf("Failed to update outbound rules: %v", err)
		} else {
			s.IpPermissionsEgressToUpdateResult = "Succeeded to update outbound rules"
		}
	}

	if len(s.TagsToDelete) > 0 {
		if _, err := client.DeleteTags(context.TODO(), &ec2.DeleteTagsInput{
			Resources: []string{
				*s.AsIsSecurityGroup.GroupId,
			},
			Tags: s.TagsToDelete,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.TagsToDeleteResult = fmt.Sprintf("Failed to delete tags: %v", err)
		} else {
			s.TagsToDeleteResult = "Succeeded to delete tags"
		}
	}

	if len(s.TagsToCreate) > 0 {
		if _, err := client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
			Resources: []string{
				*s.AsIsSecurityGroup.GroupId,
			},
			Tags: s.TagsToCreate,
		}, func(options *ec2.Options) {
			options.Region = s.RegionName
		}); err != nil {
			s.TagsToCreateResult = fmt.Sprintf("Failed to create tags: %v", err)
		} else {
			s.TagsToCreateResult = "Succeeded to create tags"
		}
	}

	log.Printf("Applied remediations")
}

func (s *SecurityGroupDelta) calculate() {
	asIsSecurityGroupIpPermissions := s.AsIsSecurityGroup.IpPermissions
	toBeSecurityGroupIpPermissions := s.ToBeSecurityGroup.IpPermissions

	s.diffIpPermissions(asIsSecurityGroupIpPermissions, toBeSecurityGroupIpPermissions, &s.IpPermissionsToRevoke, nil)
	s.diffIpPermissions(toBeSecurityGroupIpPermissions, asIsSecurityGroupIpPermissions, &s.IpPermissionsToAuthorize, &s.IpPermissionsToUpdate)

	asIsSecurityGroupIpPermissionsEgress := s.AsIsSecurityGroup.IpPermissionsEgress
	toBeSecurityGroupIpPermissionsEgress := s.ToBeSecurityGroup.IpPermissionsEgress

	s.diffIpPermissions(asIsSecurityGroupIpPermissionsEgress, toBeSecurityGroupIpPermissionsEgress, &s.IpPermissionsEgressToRevoke, nil)
	s.diffIpPermissions(toBeSecurityGroupIpPermissionsEgress, asIsSecurityGroupIpPermissionsEgress, &s.IpPermissionsEgressToAuthorize, &s.IpPermissionsEgressToUpdate)

	asIsSecurityGroupTags := s.AsIsSecurityGroup.Tags
	toBeSecurityGroupTags := s.ToBeSecurityGroup.Tags

	s.diffTags(asIsSecurityGroupTags, toBeSecurityGroupTags, &s.TagsToDelete)
	s.diffTags(toBeSecurityGroupTags, asIsSecurityGroupTags, &s.TagsToCreate)
}

func (s *SecurityGroupDelta) diffIpPermissions(thisIpPermissions []types.IpPermission, otherIpPermissions []types.IpPermission, ipPermissions *[]types.IpPermission, ipPermissionsToUpdate *[]types.IpPermission) {
	for _, thisIpPermission := range thisIpPermissions {
		ipPermissionFound := false

		for _, otherIpPermission := range otherIpPermissions {
			if determinePortRange(thisIpPermission) == determinePortRange(otherIpPermission) && determineProtocol(thisIpPermission) == determineProtocol(otherIpPermission) {
				ipPermissionFound = true

				for _, thisIpRange := range thisIpPermission.IpRanges {
					ipRangeCidrIpFound := false

					for _, otherIpRange := range otherIpPermission.IpRanges {
						if *thisIpRange.CidrIp == *otherIpRange.CidrIp {
							ipRangeCidrIpFound = true

							if ipPermissionsToUpdate != nil {
								if (thisIpRange.Description != nil && otherIpRange.Description != nil && *thisIpRange.Description != *otherIpRange.Description) ||
									(thisIpRange.Description != otherIpRange.Description && (thisIpRange.Description == nil || otherIpRange.Description == nil)) {
									*ipPermissionsToUpdate = append(*ipPermissionsToUpdate, types.IpPermission{
										FromPort:   thisIpPermission.FromPort,
										IpProtocol: thisIpPermission.IpProtocol,
										IpRanges: []types.IpRange{
											thisIpRange,
										},
										ToPort: thisIpPermission.ToPort,
									})
								}
							}

							break
						}
					}

					if !ipRangeCidrIpFound {
						*ipPermissions = append(*ipPermissions, types.IpPermission{
							FromPort:   thisIpPermission.FromPort,
							IpProtocol: thisIpPermission.IpProtocol,
							IpRanges: []types.IpRange{
								thisIpRange,
							},
							ToPort: thisIpPermission.ToPort,
						})
					}
				}

				for _, thisIpv6Range := range thisIpPermission.Ipv6Ranges {
					ipv6RangeCidrIpFound := false

					for _, otherIpv6Range := range otherIpPermission.Ipv6Ranges {
						if *thisIpv6Range.CidrIpv6 == *otherIpv6Range.CidrIpv6 {
							ipv6RangeCidrIpFound = true

							if *ipPermissionsToUpdate != nil {
								if (thisIpv6Range.Description != nil && otherIpv6Range.Description != nil && *thisIpv6Range.Description != *otherIpv6Range.Description) ||
									(thisIpv6Range.Description != otherIpv6Range.Description && (thisIpv6Range.Description == nil || otherIpv6Range.Description == nil)) {
									*ipPermissionsToUpdate = append(*ipPermissionsToUpdate, types.IpPermission{
										FromPort:   thisIpPermission.FromPort,
										IpProtocol: thisIpPermission.IpProtocol,
										Ipv6Ranges: []types.Ipv6Range{
											thisIpv6Range,
										},
										ToPort: thisIpPermission.ToPort,
									})
								}
							}

							break
						}
					}

					if !ipv6RangeCidrIpFound {
						*ipPermissions = append(*ipPermissions, types.IpPermission{
							FromPort:   thisIpPermission.FromPort,
							IpProtocol: thisIpPermission.IpProtocol,
							Ipv6Ranges: []types.Ipv6Range{
								thisIpv6Range,
							},
							ToPort: thisIpPermission.ToPort,
						})
					}
				}

				for _, thisPrefixListId := range thisIpPermission.PrefixListIds {
					prefixListIdFound := false

					for _, otherPrefixListId := range otherIpPermission.PrefixListIds {
						if *thisPrefixListId.PrefixListId == *otherPrefixListId.PrefixListId {
							prefixListIdFound = true

							if ipPermissionsToUpdate != nil && *thisPrefixListId.Description != *otherPrefixListId.Description {
								*ipPermissionsToUpdate = append(*ipPermissionsToUpdate, types.IpPermission{
									FromPort:   thisIpPermission.FromPort,
									IpProtocol: thisIpPermission.IpProtocol,
									PrefixListIds: []types.PrefixListId{
										thisPrefixListId,
									},
									ToPort: thisIpPermission.ToPort,
								})
							}

							break
						}
					}

					if !prefixListIdFound {
						*ipPermissions = append(*ipPermissions, types.IpPermission{
							FromPort:   thisIpPermission.FromPort,
							IpProtocol: thisIpPermission.IpProtocol,
							PrefixListIds: []types.PrefixListId{
								thisPrefixListId,
							},
							ToPort: thisIpPermission.ToPort,
						})
					}
				}

				for _, thisUserIdGroupPair := range thisIpPermission.UserIdGroupPairs {
					userIdGroupPairFound := false

					for _, otherUserIdGroupPair := range otherIpPermission.UserIdGroupPairs {
						if *thisUserIdGroupPair.UserId == *otherUserIdGroupPair.UserId && *thisUserIdGroupPair.GroupId == *otherUserIdGroupPair.GroupId {
							userIdGroupPairFound = true

							if *ipPermissionsToUpdate != nil && *thisUserIdGroupPair.Description != *otherUserIdGroupPair.Description {
								*ipPermissionsToUpdate = append(*ipPermissionsToUpdate, types.IpPermission{
									FromPort:   thisIpPermission.FromPort,
									IpProtocol: thisIpPermission.IpProtocol,
									ToPort:     thisIpPermission.ToPort,
									UserIdGroupPairs: []types.UserIdGroupPair{
										thisUserIdGroupPair,
									},
								})
							}

							break
						}
					}

					if !userIdGroupPairFound {
						*ipPermissions = append(*ipPermissions, types.IpPermission{
							FromPort:   thisIpPermission.FromPort,
							IpProtocol: thisIpPermission.IpProtocol,
							ToPort:     thisIpPermission.ToPort,
							UserIdGroupPairs: []types.UserIdGroupPair{
								thisUserIdGroupPair,
							},
						})
					}
				}
			}
		}

		if !ipPermissionFound {
			*ipPermissions = append(*ipPermissions, types.IpPermission{
				FromPort:         thisIpPermission.FromPort,
				IpProtocol:       thisIpPermission.IpProtocol,
				IpRanges:         thisIpPermission.IpRanges,
				Ipv6Ranges:       thisIpPermission.Ipv6Ranges,
				PrefixListIds:    thisIpPermission.PrefixListIds,
				ToPort:           thisIpPermission.ToPort,
				UserIdGroupPairs: thisIpPermission.UserIdGroupPairs,
			})
		}
	}
}

func (s *SecurityGroupDelta) diffTags(thisTags []types.Tag, otherTags []types.Tag, tags *[]types.Tag) {
	for _, thisTag := range thisTags {
		tagFound := false

		for _, otherTag := range otherTags {
			if *thisTag.Key == *otherTag.Key && *thisTag.Value == *otherTag.Value {
				tagFound = true

				break
			}
		}

		if !tagFound {
			*tags = append(*tags, types.Tag{
				Key:   thisTag.Key,
				Value: thisTag.Value,
			})
		}
	}
}

func (s *SecurityGroupDelta) tabulate() string {
	securityGroupDeltaTable := table.NewWriter()

	if s.AsIsSecurityGroup != nil {
		securityGroupDeltaTable.AppendHeader(table.Row{"As is", "To be", "Remediation", "Result"})
		securityGroupDeltaTable.SetColumnConfigs([]table.ColumnConfig{
			{
				Number:      1,
				AlignHeader: text.AlignCenter,
				Align:       text.AlignCenter,
				VAlign:      text.VAlignMiddle,
			},
			{
				Number:      2,
				AlignHeader: text.AlignCenter,
				Align:       text.AlignCenter,
				VAlign:      text.VAlignMiddle,
			},
			{
				Number:      3,
				AlignHeader: text.AlignCenter,
				Align:       text.AlignCenter,
				VAlign:      text.VAlignMiddle,
			},
			{
				Number:      4,
				AlignHeader: text.AlignCenter,
				Align:       text.AlignCenter,
				VAlign:      text.VAlignMiddle,
			},
		})
		securityGroupDeltaTable.Style().Box = table.StyleBoxRounded
		securityGroupDeltaTable.Style().Format = table.FormatOptions{
			Header: text.FormatDefault,
		}
		securityGroupDeltaTable.Style().Options.SeparateRows = true

		securityGroupDeltaTable.AppendRow(table.Row{
			tabulateSecurityGroup(*s.AsIsSecurityGroup),
			tabulateSecurityGroup(*s.ToBeSecurityGroup),
			"",
			"",
		})

		ipPermissionsRemediation := make([]string, 0, 3)
		ipPermissionsRemediationResult := make([]string, 0, 3)

		if len(s.IpPermissionsToRevoke) > 0 {
			ipPermissionsRemediation = append(ipPermissionsRemediation, tabulateIpPermissions(s.IpPermissionsToRevoke, *s.AsIsSecurityGroup, "Inbound rules to revoke"))
			ipPermissionsRemediationResult = append(ipPermissionsRemediationResult, s.IpPermissionsToRevokeResult)
		}
		if len(s.IpPermissionsToAuthorize) > 0 {
			ipPermissionsRemediation = append(ipPermissionsRemediation, tabulateIpPermissions(s.IpPermissionsToAuthorize, *s.AsIsSecurityGroup, "Inbound rules to authorize"))
			ipPermissionsRemediationResult = append(ipPermissionsRemediationResult, s.IpPermissionsToAuthorizeResult)
		}
		if len(s.IpPermissionsToUpdate) > 0 {
			ipPermissionsRemediation = append(ipPermissionsRemediation, tabulateIpPermissions(s.IpPermissionsToUpdate, *s.AsIsSecurityGroup, "Inbound rules to update"))
			ipPermissionsRemediationResult = append(ipPermissionsRemediationResult, s.IpPermissionsToUpdateResult)
		}

		securityGroupDeltaTable.AppendRow(table.Row{
			tabulateIpPermissions(s.AsIsSecurityGroup.IpPermissions, *s.AsIsSecurityGroup, "Inbound rules"),
			tabulateIpPermissions(s.ToBeSecurityGroup.IpPermissions, *s.ToBeSecurityGroup, "Inbound rules"),
			strings.Join(ipPermissionsRemediation, "\n"),
			strings.Join(ipPermissionsRemediationResult, "\n"),
		})

		ipPermissionsEgressRemediation := make([]string, 0, 3)
		ipPermissionsEgressRemediationResult := make([]string, 0, 3)

		if len(s.IpPermissionsEgressToRevoke) > 0 {
			ipPermissionsEgressRemediation = append(ipPermissionsEgressRemediation, tabulateIpPermissions(s.IpPermissionsEgressToRevoke, *s.AsIsSecurityGroup, "Outbound rules to revoke"))
			ipPermissionsEgressRemediationResult = append(ipPermissionsEgressRemediationResult, s.IpPermissionsEgressToRevokeResult)
		}
		if len(s.IpPermissionsEgressToAuthorize) > 0 {
			ipPermissionsEgressRemediation = append(ipPermissionsEgressRemediation, tabulateIpPermissions(s.IpPermissionsEgressToAuthorize, *s.AsIsSecurityGroup, "Outbound rules to authorize"))
			ipPermissionsEgressRemediationResult = append(ipPermissionsEgressRemediationResult, s.IpPermissionsEgressToAuthorizeResult)
		}
		if len(s.IpPermissionsEgressToUpdate) > 0 {
			ipPermissionsEgressRemediation = append(ipPermissionsEgressRemediation, tabulateIpPermissions(s.IpPermissionsEgressToUpdate, *s.AsIsSecurityGroup, "Outbound rules to update"))
			ipPermissionsEgressRemediationResult = append(ipPermissionsEgressRemediationResult, s.IpPermissionsEgressToUpdateResult)
		}

		securityGroupDeltaTable.AppendRow(table.Row{
			tabulateIpPermissions(s.AsIsSecurityGroup.IpPermissionsEgress, *s.AsIsSecurityGroup, "Outbound rules"),
			tabulateIpPermissions(s.ToBeSecurityGroup.IpPermissionsEgress, *s.ToBeSecurityGroup, "Outbound rules"),
			strings.Join(ipPermissionsEgressRemediation, "\n"),
			strings.Join(ipPermissionsEgressRemediationResult, "\n"),
		})

		tagsRemediation := make([]string, 0, 3)
		tagsRemediationResult := make([]string, 0, 3)

		if len(s.TagsToDelete) > 0 {
			tagsRemediation = append(tagsRemediation, tabulateTags(s.TagsToDelete, "Tags to delete"))
			tagsRemediationResult = append(tagsRemediationResult, s.TagsToDeleteResult)
		}
		if len(s.TagsToCreate) > 0 {
			tagsRemediation = append(tagsRemediation, tabulateTags(s.TagsToCreate, "Tags to create"))
			tagsRemediationResult = append(tagsRemediationResult, s.TagsToCreateResult)
		}

		securityGroupDeltaTable.AppendRow(table.Row{
			tabulateTags(s.AsIsSecurityGroup.Tags, "Tags"),
			tabulateTags(s.ToBeSecurityGroup.Tags, "Tags"),
			strings.Join(tagsRemediation, "\n"),
			strings.Join(tagsRemediationResult, "\n"),
		})
	} else {
		securityGroupDeltaTable.AppendHeader(table.Row{"As is", "To be"})
		securityGroupDeltaTable.SetColumnConfigs([]table.ColumnConfig{
			{
				Number:      1,
				AlignHeader: text.AlignCenter,
				Align:       text.AlignCenter,
			},
			{
				Number:      2,
				AlignHeader: text.AlignCenter,
				Align:       text.AlignCenter,
			},
		})
		securityGroupDeltaTable.Style().Box = table.StyleBoxRounded
		securityGroupDeltaTable.Style().Format = table.FormatOptions{
			Header: text.FormatDefault,
		}
		securityGroupDeltaTable.Style().Options.SeparateRows = true

		securityGroupDeltaTable.AppendRow(table.Row{
			fmt.Sprintf("No matching security group found with ID: %s in VPC: %s", *s.ToBeSecurityGroup.GroupId, *s.ToBeSecurityGroup.VpcId),
			tabulateSecurityGroup(*s.ToBeSecurityGroup),
		})

		securityGroupDeltaTable.AppendRow(table.Row{
			"",
			tabulateIpPermissions(s.ToBeSecurityGroup.IpPermissions, *s.ToBeSecurityGroup, "Inbound rules"),
		})

		securityGroupDeltaTable.AppendRow(table.Row{
			"",
			tabulateIpPermissions(s.ToBeSecurityGroup.IpPermissionsEgress, *s.ToBeSecurityGroup, "Outbound rules"),
		})

		securityGroupDeltaTable.AppendRow(table.Row{
			"",
			tabulateTags(s.ToBeSecurityGroup.Tags, "Tags"),
		})
	}

	return securityGroupDeltaTable.Render()
}
