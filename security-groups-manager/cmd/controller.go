package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Controller struct {
	Client                         *ec2.Client
	SecurityGroupIdRegionNameMutex sync.Mutex
	SecurityGroupIdRegionName      map[string]string
	AsIsSecurityGroups             []types.SecurityGroup
	ToBeSecurityGroups             []types.SecurityGroup
	SecurityGroupDeltas            []SecurityGroupDelta
}

func NewController(client *ec2.Client) *Controller {
	controller := new(Controller)

	controller.Client = client
	controller.SecurityGroupIdRegionName = make(map[string]string)
	controller.AsIsSecurityGroups = make([]types.SecurityGroup, 0)
	controller.ToBeSecurityGroups = make([]types.SecurityGroup, 0)
	controller.SecurityGroupDeltas = make([]SecurityGroupDelta, 0)

	return controller
}

func (c *Controller) CalculateSecurityGroupDeltas() {
	log.Printf("Calculating security group deltas")

	securityGroupDeltaChannel := make(chan SecurityGroupDelta)

	for _, toBeSecurityGroup := range c.ToBeSecurityGroups {
		go func(toBeSecurityGroup types.SecurityGroup) {
			securityGroupDelta := NewSecurityGroupDelta(&toBeSecurityGroup)

			for _, asIsSecurityGroup := range c.AsIsSecurityGroups {
				if *toBeSecurityGroup.VpcId == *asIsSecurityGroup.VpcId && *toBeSecurityGroup.GroupId == *asIsSecurityGroup.GroupId {
					securityGroupDelta.AsIsSecurityGroup = &asIsSecurityGroup

					c.SecurityGroupIdRegionNameMutex.Lock()
					securityGroupDelta.RegionName = c.SecurityGroupIdRegionName[*securityGroupDelta.AsIsSecurityGroup.GroupId]
					c.SecurityGroupIdRegionNameMutex.Unlock()

					securityGroupDelta.calculate()

					break
				}
			}

			securityGroupDeltaChannel <- *securityGroupDelta
		}(toBeSecurityGroup)
	}

	for range c.ToBeSecurityGroups {
		securityGroupDelta := <-securityGroupDeltaChannel

		c.SecurityGroupDeltas = append(c.SecurityGroupDeltas, securityGroupDelta)
	}

	log.Printf("Calculated security group deltas")
}

func (c *Controller) InitAsIsSecurityGroups() error {
	describeRegionsOutput, err := c.Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		log.Printf("Unable to describe regions: %v", err)

		return err
	}

	asIsSecurityGroupsChannel := make(chan []types.SecurityGroup)

	for _, region := range describeRegionsOutput.Regions {
		if *region.OptInStatus != "not-opted-in" {
			go func(regionName string) {
				describeSecurityGroupsOutput, err := c.Client.DescribeSecurityGroups(context.TODO(), nil, func(options *ec2.Options) {
					options.Region = regionName
				})
				if err != nil {
					log.Printf("Unable to describe security groups in region %s: %v", regionName, err)

					asIsSecurityGroupsChannel <- nil

					return
				}

				for _, securityGroup := range describeSecurityGroupsOutput.SecurityGroups {
					c.SecurityGroupIdRegionNameMutex.Lock()
					c.SecurityGroupIdRegionName[*securityGroup.GroupId] = regionName
					c.SecurityGroupIdRegionNameMutex.Unlock()
				}

				asIsSecurityGroupsChannel <- describeSecurityGroupsOutput.SecurityGroups
			}(*region.RegionName)
		}
	}

	for _, region := range describeRegionsOutput.Regions {
		if *region.OptInStatus != "not-opted-in" {
			asIsSecurityGroups := <-asIsSecurityGroupsChannel

			if len(asIsSecurityGroups) > 0 {
				c.AsIsSecurityGroups = append(c.AsIsSecurityGroups, asIsSecurityGroups...)
			}
		}
	}

	return nil
}

func (c *Controller) InitToBeSecurityGroups(configuration *Configuration) {
	toBeSecurityGroupChannel := make(chan *types.SecurityGroup)

	for _, configuredSecurityGroup := range configuration.SecurityGroups {
		go func(configuredSecurityGroup SecurityGroup) {
			configuredSecurityGroup.consolidateHostsAndIpRanges(configuredSecurityGroup.IpPermissions)
			configuredSecurityGroup.consolidateHostsAndIpRanges(configuredSecurityGroup.IpPermissionsEgress)

			var toBeSecurityGroup types.SecurityGroup

			b, err := json.Marshal(configuredSecurityGroup)
			if err != nil {
				log.Printf("Unable to marshal configured security group %s: %v", *configuredSecurityGroup.GroupName, err)

				toBeSecurityGroupChannel <- nil

				return
			}
			if err := json.Unmarshal(b, &toBeSecurityGroup); err != nil {
				log.Printf("Unable to unmarshal security group %s: %v", *configuredSecurityGroup.GroupName, err)

				toBeSecurityGroupChannel <- nil

				return
			}

			toBeSecurityGroupChannel <- &toBeSecurityGroup
		}(configuredSecurityGroup)
	}

	for range configuration.SecurityGroups {
		toBeSecurityGroup := <-toBeSecurityGroupChannel
		if toBeSecurityGroup != nil {
			c.ToBeSecurityGroups = append(c.ToBeSecurityGroups, *toBeSecurityGroup)
		}
	}
}

func (c *Controller) ProcessSecurityGroupDeltas() {
	log.Printf("Processing security group deltas")

	securityGroupDeltaApplyChannel := make(chan SecurityGroupDelta)

	for _, securityGroupDelta := range c.SecurityGroupDeltas {
		go func(securityGroupDelta SecurityGroupDelta) {
			if securityGroupDelta.AsIsSecurityGroup != nil && (len (securityGroupDelta.IpPermissionsToAuthorize) > 0 || len(securityGroupDelta.IpPermissionsToRevoke) > 0 || len(securityGroupDelta.IpPermissionsToUpdate) > 0 ||
				len(securityGroupDelta.IpPermissionsEgressToAuthorize) > 0 || len(securityGroupDelta.IpPermissionsEgressToRevoke) > 0 || len(securityGroupDelta.IpPermissionsEgressToUpdate) > 0 ||
				len(securityGroupDelta.TagsToCreate) > 0 || len(securityGroupDelta.TagsToDelete) > 0) {
				securityGroupDelta.apply(c.Client)
			}

			securityGroupDeltaApplyChannel <- securityGroupDelta
		}(securityGroupDelta)
	}

	for range c.SecurityGroupDeltas {
		securityGroupDelta := <-securityGroupDeltaApplyChannel

		if securityGroupDelta.AsIsSecurityGroup == nil || len (securityGroupDelta.IpPermissionsToAuthorize) > 0 || len(securityGroupDelta.IpPermissionsToRevoke) > 0 || len(securityGroupDelta.IpPermissionsToUpdate) > 0 ||
				len(securityGroupDelta.IpPermissionsEgressToAuthorize) > 0 || len(securityGroupDelta.IpPermissionsEgressToRevoke) > 0 || len(securityGroupDelta.IpPermissionsEgressToUpdate) > 0 ||
				len(securityGroupDelta.TagsToCreate) > 0 || len(securityGroupDelta.TagsToDelete) > 0 {
			log.Println("\n" + securityGroupDelta.tabulate())
		} else {
			log.Printf("%s / %s is up to date", *securityGroupDelta.AsIsSecurityGroup.GroupId, *securityGroupDelta.AsIsSecurityGroup.GroupName)
		}
	}

	log.Printf("Processed security group deltas")
}
