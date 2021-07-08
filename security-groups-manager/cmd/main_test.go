package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

var client *ec2.Client

var regionNameSecurityGroupMutex sync.Mutex
var regionNameSecurityGroup = map[string]types.SecurityGroup{}

var securityGroupIdRegionNameMutex sync.Mutex
var securityGroupIdRegionName = map[string]string{}

func init() {
	awsConfiguration, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	client = ec2.NewFromConfig(awsConfiguration)
}

func extractRegion(groupName string) string {
	return strings.Split(groupName, "_")[1]
}

func generateConfigurationFromTemplate(t *testing.T, templateFile string) string {
	securityGroups := make([]types.SecurityGroup, 0, len(regionNameSecurityGroup))

	for _, securityGroup := range regionNameSecurityGroup {
		securityGroups = append(securityGroups, securityGroup)
	}

	b, err := os.ReadFile(templateFile)
	if err != nil {
		t.Errorf("Unable to read template setup.json: %v", err)
	}

	templateString := &strings.Builder{}

	template := template.Must(template.New("configuration").Funcs(template.FuncMap{"extractRegion": extractRegion}).Parse(string(b)))
	template.Execute(templateString, securityGroups)

	return templateString.String()
}

func randomizeRegions() []types.Region {
	describeRegionsOutput, err := client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Unable to describe regions: %v", err)
	}

	regions := make([]types.Region, 0, len(describeRegionsOutput.Regions))

	for _, region := range describeRegionsOutput.Regions {
		if *region.OptInStatus != "not-opted-in" {
			regions = append(regions, region)
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(regions), func(i, j int) {
		regions[i], regions[j] = regions[j], regions[i]
	})

	// TODO: Convert hardcoded limit to an environment variable
	return regions[:2]
}

func runTemplateBasedTest(t *testing.T, templateFile string) bool {
	runNextTest := true

	os.Setenv("CONFIGURATION", generateConfigurationFromTemplate(t, templateFile))
	os.Setenv("DEBUG", "false")

	controller, err := execute()
	if err != nil {
		runNextTest = false

		t.Error("Unexpected error encountered")

		return runNextTest
	}

	for _, toBeSecurityGroup := range controller.ToBeSecurityGroups {
		regionName := securityGroupIdRegionName[*toBeSecurityGroup.GroupId]

		describeSecurityGroupsOutput, err := client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
			GroupIds: []string{*toBeSecurityGroup.GroupId},
		}, func(options *ec2.Options) {
			options.Region = regionName
		})
		if err != nil {
			runNextTest = false

			t.Errorf("Unable to describe security groups in region %s: %v", regionName, err)

			return runNextTest
		}

		if len(describeSecurityGroupsOutput.SecurityGroups) != 1 {
			runNextTest = false

			t.Error("Unexpected error encountered")

			return runNextTest
		}

		sortSecurityGroup(toBeSecurityGroup)
		sortSecurityGroup(describeSecurityGroupsOutput.SecurityGroups[0])

		if diff := deep.Equal(toBeSecurityGroup, describeSecurityGroupsOutput.SecurityGroups[0]); diff != nil {
			runNextTest = false

			t.Errorf("Want != Got: %v", diff)

			return runNextTest
		}
	}

	return runNextTest
}

func setup() bool {
	ok := true

	regions := randomizeRegions()
	setupOkChannel := make(chan bool)

	for _, region := range regions {
		go func(regionName string) {
			log.Printf("Creating security group in %s\n", regionName)

			createSecurityGroupOutput, err := client.CreateSecurityGroup(context.TODO(), &ec2.CreateSecurityGroupInput{
				Description: aws.String("Security Group created by SecurityGroupsManager test suite in " + regionName),
				GroupName:   aws.String("SecurityGroupsManager_" + regionName + "_SG"),
			}, func(options *ec2.Options) {
				options.Region = regionName
			})
			if err != nil {
				log.Printf("Unable to create security group in %s: %v\n", regionName, err)

				setupOkChannel <- false

				return
			}

			log.Printf("Created security group in %s\n", regionName)

			regionNameSecurityGroupMutex.Lock()
			regionNameSecurityGroup[regionName] = types.SecurityGroup{
				GroupId: createSecurityGroupOutput.GroupId,
			}
			regionNameSecurityGroupMutex.Unlock()

			describeSecurityGroupsOutput, err := client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
				GroupIds: []string{*createSecurityGroupOutput.GroupId},
			}, func(options *ec2.Options) {
				options.Region = regionName
			})
			if err != nil {
				log.Printf("Unable to describe security groups in region %s: %v", regionName, err)

				setupOkChannel <- false

				return
			} else {
				regionNameSecurityGroupMutex.Lock()
				regionNameSecurityGroup[regionName] = describeSecurityGroupsOutput.SecurityGroups[0]
				regionNameSecurityGroupMutex.Unlock()

			}

			securityGroupIdRegionNameMutex.Lock()
			securityGroupIdRegionName[*createSecurityGroupOutput.GroupId] = regionName
			securityGroupIdRegionNameMutex.Unlock()

			setupOkChannel <- true
		}(*region.RegionName)
	}

	for range regions {
		if !<-setupOkChannel {
			ok = false
		}
	}

	return ok
}

func sortSecurityGroup(securityGroup types.SecurityGroup) {
	for _, ipPermission := range securityGroup.IpPermissions {
		sort.SliceStable(ipPermission.IpRanges, func(i int, j int) bool {
			if *ipPermission.IpRanges[i].CidrIp < *ipPermission.IpRanges[j].CidrIp {
				return true
			}
			if *ipPermission.IpRanges[i].CidrIp > *ipPermission.IpRanges[j].CidrIp {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermission.IpRanges[i].Description != nil {
				iDescription = *ipPermission.IpRanges[i].Description
			}
			if ipPermission.IpRanges[j].Description != nil {
				jDescription = *ipPermission.IpRanges[j].Description
			}
			return iDescription < jDescription
		})

		sort.SliceStable(ipPermission.Ipv6Ranges, func(i int, j int) bool {
			if *ipPermission.Ipv6Ranges[i].CidrIpv6 < *ipPermission.Ipv6Ranges[j].CidrIpv6 {
				return true
			}
			if *ipPermission.Ipv6Ranges[i].CidrIpv6 > *ipPermission.Ipv6Ranges[j].CidrIpv6 {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermission.Ipv6Ranges[i].Description != nil {
				iDescription = *ipPermission.Ipv6Ranges[i].Description
			}
			if ipPermission.Ipv6Ranges[j].Description != nil {
				jDescription = *ipPermission.Ipv6Ranges[j].Description
			}
			return iDescription < jDescription
		})

		sort.SliceStable(ipPermission.PrefixListIds, func(i int, j int) bool {
			if *ipPermission.PrefixListIds[i].PrefixListId < *ipPermission.PrefixListIds[j].PrefixListId {
				return true
			}
			if *ipPermission.PrefixListIds[i].PrefixListId > *ipPermission.PrefixListIds[j].PrefixListId {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermission.PrefixListIds[i].Description != nil {
				iDescription = *ipPermission.PrefixListIds[i].Description
			}
			if ipPermission.PrefixListIds[j].Description != nil {
				jDescription = *ipPermission.PrefixListIds[j].Description
			}
			return iDescription < jDescription
		})

		sort.SliceStable(ipPermission.UserIdGroupPairs, func(i int, j int) bool {
			if *ipPermission.UserIdGroupPairs[i].GroupId < *ipPermission.UserIdGroupPairs[j].GroupId {
				return true
			}
			if *ipPermission.UserIdGroupPairs[i].GroupId > *ipPermission.UserIdGroupPairs[j].GroupId {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermission.UserIdGroupPairs[i].Description != nil {
				iDescription = *ipPermission.UserIdGroupPairs[i].Description
			}
			if ipPermission.UserIdGroupPairs[j].Description != nil {
				jDescription = *ipPermission.UserIdGroupPairs[j].Description
			}
			return iDescription < jDescription
		})
	}

	for _, ipPermissionEgress := range securityGroup.IpPermissionsEgress {
		sort.SliceStable(ipPermissionEgress.IpRanges, func(i int, j int) bool {
			if *ipPermissionEgress.IpRanges[i].CidrIp < *ipPermissionEgress.IpRanges[j].CidrIp {
				return true
			}
			if *ipPermissionEgress.IpRanges[i].CidrIp > *ipPermissionEgress.IpRanges[j].CidrIp {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermissionEgress.IpRanges[i].Description != nil {
				iDescription = *ipPermissionEgress.IpRanges[i].Description
			}
			if ipPermissionEgress.IpRanges[j].Description != nil {
				jDescription = *ipPermissionEgress.IpRanges[j].Description
			}
			return iDescription < jDescription
		})

		sort.SliceStable(ipPermissionEgress.Ipv6Ranges, func(i int, j int) bool {
			if *ipPermissionEgress.Ipv6Ranges[i].CidrIpv6 < *ipPermissionEgress.Ipv6Ranges[j].CidrIpv6 {
				return true
			}
			if *ipPermissionEgress.Ipv6Ranges[i].CidrIpv6 > *ipPermissionEgress.Ipv6Ranges[j].CidrIpv6 {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermissionEgress.Ipv6Ranges[i].Description != nil {
				iDescription = *ipPermissionEgress.Ipv6Ranges[i].Description
			}
			if ipPermissionEgress.Ipv6Ranges[j].Description != nil {
				jDescription = *ipPermissionEgress.Ipv6Ranges[j].Description
			}
			return iDescription < jDescription
		})

		sort.SliceStable(ipPermissionEgress.PrefixListIds, func(i int, j int) bool {
			if *ipPermissionEgress.PrefixListIds[i].PrefixListId < *ipPermissionEgress.PrefixListIds[j].PrefixListId {
				return true
			}
			if *ipPermissionEgress.PrefixListIds[i].PrefixListId > *ipPermissionEgress.PrefixListIds[j].PrefixListId {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermissionEgress.PrefixListIds[i].Description != nil {
				iDescription = *ipPermissionEgress.PrefixListIds[i].Description
			}
			if ipPermissionEgress.PrefixListIds[j].Description != nil {
				jDescription = *ipPermissionEgress.PrefixListIds[j].Description
			}
			return iDescription < jDescription
		})

		sort.SliceStable(ipPermissionEgress.UserIdGroupPairs, func(i int, j int) bool {
			if *ipPermissionEgress.UserIdGroupPairs[i].GroupId < *ipPermissionEgress.UserIdGroupPairs[j].GroupId {
				return true
			}
			if *ipPermissionEgress.UserIdGroupPairs[i].GroupId > *ipPermissionEgress.UserIdGroupPairs[j].GroupId {
				return false
			}

			iDescription := ""
			jDescription := ""

			if ipPermissionEgress.UserIdGroupPairs[i].Description != nil {
				iDescription = *ipPermissionEgress.UserIdGroupPairs[i].Description
			}
			if ipPermissionEgress.UserIdGroupPairs[j].Description != nil {
				jDescription = *ipPermissionEgress.UserIdGroupPairs[j].Description
			}
			return iDescription < jDescription
		})
	}

	sort.SliceStable(securityGroup.IpPermissions, func(i int, j int) bool {
		if *securityGroup.IpPermissions[i].IpProtocol < *securityGroup.IpPermissions[j].IpProtocol {
			return true
		}
		if *securityGroup.IpPermissions[i].IpProtocol > *securityGroup.IpPermissions[j].IpProtocol {
			return false
		}

		iFromPort := int32(-65536)
		jFromPort := int32(-65536)

		if securityGroup.IpPermissions[i].FromPort != nil {
			iFromPort = *securityGroup.IpPermissions[i].FromPort
		}
		if securityGroup.IpPermissions[j].FromPort != nil {
			jFromPort = *securityGroup.IpPermissions[j].FromPort
		}

		if iFromPort < jFromPort {
			return true
		}
		if iFromPort > jFromPort {
			return false
		}

		iToPort := int32(-65536)
		jToPort := int32(-65536)

		if securityGroup.IpPermissions[i].ToPort != nil {
			iToPort = *securityGroup.IpPermissions[i].ToPort
		}
		if securityGroup.IpPermissions[j].ToPort != nil {
			jToPort = *securityGroup.IpPermissions[j].ToPort
		}

		return iToPort < jToPort
	})

	sort.SliceStable(securityGroup.IpPermissionsEgress, func(i int, j int) bool {
		if *securityGroup.IpPermissionsEgress[i].IpProtocol < *securityGroup.IpPermissions[j].IpProtocol {
			return true
		}
		if *securityGroup.IpPermissionsEgress[i].IpProtocol > *securityGroup.IpPermissions[j].IpProtocol {
			return false
		}

		iFromPort := int32(-65536)
		jFromPort := int32(-65536)

		if securityGroup.IpPermissionsEgress[i].FromPort != nil {
			iFromPort = *securityGroup.IpPermissionsEgress[i].FromPort
		}
		if securityGroup.IpPermissions[j].FromPort != nil {
			jFromPort = *securityGroup.IpPermissions[j].FromPort
		}

		if iFromPort < jFromPort {
			return true
		}
		if iFromPort > jFromPort {
			return false
		}

		iToPort := int32(-65536)
		jToPort := int32(-65536)

		if securityGroup.IpPermissionsEgress[i].ToPort != nil {
			iToPort = *securityGroup.IpPermissionsEgress[i].ToPort
		}
		if securityGroup.IpPermissions[j].ToPort != nil {
			jToPort = *securityGroup.IpPermissions[j].ToPort
		}

		return iToPort < jToPort
	})

	sort.SliceStable(securityGroup.Tags, func(i int, j int) bool {
		return *securityGroup.Tags[i].Key < *securityGroup.Tags[j].Key
	})
}

func teardown() bool {
	ok := true

	p := recover()

	teardownOkChannel := make(chan bool)

	for regionName, securityGroup := range regionNameSecurityGroup {
		go func(regionName string, securityGroup types.SecurityGroup) {
			log.Printf("Deleting security group in %s\n", regionName)

			_, err := client.DeleteSecurityGroup(context.TODO(), &ec2.DeleteSecurityGroupInput{
				GroupId: securityGroup.GroupId,
			}, func(options *ec2.Options) {
				options.Region = regionName
			})
			if err != nil {
				log.Printf("Unable to delete security group in %s: %v\n", regionName, err)

				teardownOkChannel <- false

				return
			}

			log.Printf("Deleted security group in %s\n", regionName)

			teardownOkChannel <- true
		}(regionName, securityGroup)
	}

	for range regionNameSecurityGroup {
		if !<-teardownOkChannel {
			ok = false
		}
	}

	if p != nil {
		panic(p)
	}

	return ok
}

func TestSecurityGroupsManager(t *testing.T) {
	runNextTest := true

	t.Run("Step #1 (Authorize rules + Create tags)", func(t *testing.T) {
		runNextTest = runTemplateBasedTest(t, "../testdata/step_1.json")
	})
	t.Run("Step #2 (Update rules + Update tags)", func(t *testing.T) {
		if runNextTest {
			runNextTest = runTemplateBasedTest(t, "../testdata/step_2.json")
		}
	})
	t.Run("Step #3.1 (Revoke rules + Delete tags)", func(t *testing.T) {
		if runNextTest {
			runNextTest = runTemplateBasedTest(t, "../testdata/step_3.json")
		}
	})
	t.Run("Step #3.2 (No change)", func(t *testing.T) {
		if runNextTest {
			runNextTest = runTemplateBasedTest(t, "../testdata/step_3.json")
		}
	})
	t.Run("Step #4 (Nonexistent)", func(t *testing.T) {
		if runNextTest {
			b, err := os.ReadFile("../testdata/step_4.json")
			if err != nil {
				t.Errorf("Unable to read template setup.json: %v", err)
			}

			os.Setenv("CONFIGURATION", string(b))
			os.Setenv("DEBUG", "false")

			controller, err := execute()
			if err != nil {
				runNextTest = false

				t.Error("Unexpected error encountered")
			}

			assert.Nil(t, controller.SecurityGroupDeltas[0].AsIsSecurityGroup)
		}
	})
}

func TestMain(m *testing.M) {
	defer teardown()

	if ok := setup(); ok {
		m.Run()
	}
}
