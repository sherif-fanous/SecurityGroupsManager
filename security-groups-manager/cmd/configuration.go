package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"inet.af/netaddr"
)

type Configuration struct {
	SecurityGroups []SecurityGroup
}

func NewConfiguration(marshaledConfiguration string) (*Configuration, error) {
	configuration := new(Configuration)

	debugf("Unmarshalling configuration")

	if err := json.Unmarshal([]byte(marshaledConfiguration), configuration); err != nil {
		log.Printf("Unable to unmarshal configuration: %v", err)

		return nil, err
	}

	debugf("Unmarshalled configuration")

	return configuration, nil
}

type SecurityGroup struct {
	Description         *string
	GroupId             *string
	GroupName           *string
	IpPermissions       []IpPermission
	IpPermissionsEgress []IpPermission
	OwnerId             *string
	Tags                []types.Tag
	VpcId               *string
}

func (s *SecurityGroup) consolidateHostsAndIpRanges(ipPermissions []IpPermission) {
	for i := range ipPermissions {
		configuredIpPermission := &ipPermissions[i]

		for _, host := range configuredIpPermission.Hosts {
			addresses, err := net.LookupHost(*host.FQDN)
			if err != nil {
				log.Printf("Unable to lookup host: %v", err)
			}

			for _, address := range addresses {
				ip, err := netaddr.ParseIP(address)
				if err != nil {
					log.Printf("Host %s resolved to %s: %v", *host.FQDN, address, err)

					continue
				}

				if ip.Is6() {
					cidrIpv6 := address + "/128"
					cidrIpv6Found := false

					for _, Ipv6Range := range configuredIpPermission.Ipv6Ranges {
						if cidrIpv6 == *Ipv6Range.CidrIpv6 {
							cidrIpv6Found = true

							break
						}
					}

					if !cidrIpv6Found {
						if configuredIpPermission.Ipv6Ranges == nil {
							configuredIpPermission.Ipv6Ranges = make([]types.Ipv6Range, 0, 1)
						}

						configuredIpPermission.Ipv6Ranges = append(configuredIpPermission.Ipv6Ranges, types.Ipv6Range{
							CidrIpv6:    &cidrIpv6,
							Description: host.Description,
						})
					}
				} else {
					cidrIp := address + "/32"
					cidrIpFound := false

					for _, IpRange := range configuredIpPermission.IpRanges {
						if cidrIp == *IpRange.CidrIp {
							cidrIpFound = true

							break
						}
					}

					if !cidrIpFound {
						if configuredIpPermission.IpRanges == nil {
							configuredIpPermission.IpRanges = make([]types.IpRange, 0, 1)
						}

						configuredIpPermission.IpRanges = append(configuredIpPermission.IpRanges, types.IpRange{
							CidrIp:      &cidrIp,
							Description: host.Description,
						})
					}
				}
			}
		}
	}
}

type IpPermission struct {
	FromPort         *int32
	Hosts            []Host
	IpProtocol       *string
	IpRanges         []types.IpRange
	Ipv6Ranges       []types.Ipv6Range
	PrefixListIds    []types.PrefixListId
	ToPort           *int32
	UserIdGroupPairs []types.UserIdGroupPair
}

type Host struct {
	FQDN        *string
	Description *string
}
