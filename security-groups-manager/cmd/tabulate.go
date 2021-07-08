package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var icmpTypeNumbers = map[int]string{
	0:   "Echo Reply",
	1:   "Unassigned",
	2:   "Unassigned",
	3:   "Destination Unreachable",
	4:   "Source Quench (Deprecated)",
	5:   "Redirect",
	6:   "Alternate Host Address (Deprecated)",
	7:   "Unassigned",
	8:   "Echo",
	9:   "Router Advertisement",
	10:  "Router Solicitation",
	11:  "Time Exceeded",
	12:  "Parameter Problem",
	13:  "Timestamp",
	14:  "Timestamp Reply",
	15:  "Information Request (Deprecated)",
	16:  "Information Reply (Deprecated)",
	17:  "Address Mask Request (Deprecated)",
	18:  "Address Mask Reply (Deprecated)",
	19:  "Reserved (for Security)",
	30:  "Traceroute (Deprecated)",
	31:  "Datagram Conversion Error (Deprecated)",
	32:  "Mobile Host Redirect (Deprecated)",
	33:  "IPv6 Where-Are-You (Deprecated)",
	34:  "IPv6 I-Am-Here (Deprecated)",
	35:  "Mobile Registration Request (Deprecated)",
	36:  "Mobile Registration Reply (Deprecated)",
	37:  "Domain Name Request (Deprecated)",
	38:  "Domain Name Reply (Deprecated)",
	39:  "SKIP (Deprecated)",
	40:  "Photuris",
	41:  "ICMP messages utilized by experimental mobility protocols such as Seamoby",
	42:  "Extended Echo Request",
	43:  "Extended Echo Reply",
	253: "RFC3692-style Experiment 1",
	254: "RFC3692-style Experiment 2",
	255: "Reserved",
}

var protocols = map[string]string{
	"-1":     "All",
	"0":      "HOPOPT",
	"1":      "ICMP",
	"2":      "IGMP",
	"3":      "GGP",
	"4":      "IPv4",
	"5":      "ST",
	"6":      "TCP",
	"7":      "CBT",
	"8":      "EGP",
	"9":      "IGP",
	"10":     "BBN-RCC-MON",
	"11":     "NVP-II",
	"12":     "PUP",
	"13":     "ARGUS (deprecated)",
	"14":     "EMCON",
	"15":     "XNET",
	"16":     "CHAOS",
	"17":     "UDP",
	"18":     "MUX",
	"19":     "DCN-MEAS",
	"20":     "HMP",
	"21":     "PRM",
	"22":     "XNS-IDP",
	"23":     "TRUNK-1",
	"24":     "TRUNK-2",
	"25":     "LEAF-1",
	"26":     "LEAF-2",
	"27":     "RDP",
	"28":     "IRTP",
	"29":     "ISO-TP4",
	"30":     "NETBLT",
	"31":     "MFE-NSP",
	"32":     "MERIT-INP",
	"33":     "DCCP",
	"34":     "3PC",
	"35":     "IDPR",
	"36":     "XTP",
	"37":     "DDP",
	"38":     "IDPR-CMTP",
	"39":     "TP++",
	"40":     "IL",
	"41":     "IPv6",
	"42":     "SDRP",
	"43":     "IPv6-Route",
	"44":     "IPv6-Frag",
	"45":     "IDRP",
	"46":     "RSVP",
	"47":     "GRE",
	"48":     "DSR",
	"49":     "BNA",
	"50":     "ESP",
	"51":     "AH",
	"52":     "I-NLSP",
	"53":     "SWIPE (deprecated)",
	"54":     "NARP",
	"55":     "MOBILE",
	"56":     "TLSP",
	"57":     "SKIP",
	"58":     "IPv6-ICMP",
	"59":     "IPv6-NoNxt",
	"60":     "IPv6-Opts",
	"62":     "CFTP",
	"64":     "SAT-EXPAK",
	"65":     "KRYPTOLAN",
	"66":     "RVD",
	"67":     "IPPC",
	"69":     "SAT-MON",
	"70":     "VISA",
	"71":     "IPCV",
	"72":     "CPNX",
	"73":     "CPHB",
	"74":     "WSN",
	"75":     "PVP",
	"76":     "BR-SAT-MON",
	"77":     "SUN-ND",
	"78":     "WB-MON",
	"79":     "WB-EXPAK",
	"80":     "ISO-IP",
	"81":     "VMTP",
	"82":     "SECURE-VMTP",
	"83":     "VINES",
	"84":     "TTP / IPTM",
	"85":     "NSFNET-IGP",
	"86":     "DGP",
	"87":     "TCF",
	"88":     "EIGRP",
	"89":     "OSPFIGP",
	"90":     "Sprite-RPC",
	"91":     "LARP",
	"92":     "MTP",
	"93":     "AX.25",
	"94":     "IPIP",
	"95":     "MICP (deprecated)",
	"96":     "SCC-SP",
	"97":     "ETHERIP",
	"98":     "ENCAP",
	"100":    "GMTP",
	"101":    "IFMP",
	"102":    "PNNI",
	"103":    "PIM",
	"104":    "ARIS",
	"105":    "SCPS",
	"106":    "QNX",
	"107":    "A/N",
	"108":    "IPComp",
	"109":    "SNP",
	"110":    "Compaq-Peer",
	"111":    "IPX-in-IP",
	"112":    "VRRP",
	"113":    "PGM",
	"115":    "L2TP",
	"116":    "DDX",
	"117":    "IATP",
	"118":    "STP",
	"119":    "SRP",
	"120":    "UTI",
	"121":    "SMP",
	"122":    "SM (deprecated)",
	"123":    "PTP",
	"124":    "ISIS over IPv4",
	"125":    "FIRE",
	"126":    "CRTP",
	"127":    "CRUDP",
	"128":    "SSCOPMCE",
	"129":    "IPLT",
	"130":    "SPS",
	"131":    "PIPE",
	"132":    "SCTP",
	"133":    "FC",
	"134":    "RSVP-E2E-IGNORE",
	"135":    "Mobility Header",
	"136":    "UDPLite",
	"137":    "MPLS-in-IP",
	"138":    "manet",
	"139":    "HIP",
	"140":    "Shim6",
	"141":    "WESP",
	"142":    "ROHC",
	"143":    "Ethernet",
	"255":    "Reserved",
	"icmp":   "ICMP",
	"icmpv6": "IPv6-ICMP",
	"tcp":    "TCP",
	"udp":    "UDP",
}

func determinePortRange(ipPermission types.IpPermission) string {
	portRange := "All"

	if ipPermission.FromPort != nil && *ipPermission.FromPort != -1 {
		if *ipPermission.IpProtocol == "icmp" || *ipPermission.IpProtocol == "icmpv6" {
			var ok bool

			portRange, ok = icmpTypeNumbers[int(*ipPermission.FromPort)]
			if !ok {
				portRange = "Unknown"
			}
		} else {
			portRange = strconv.Itoa(int(*ipPermission.FromPort))
			if ipPermission.ToPort != nil && *ipPermission.FromPort != *ipPermission.ToPort {
				portRange += " - " + strconv.Itoa(int(*ipPermission.ToPort))
			}
		}
	}

	return portRange
}

func determineProtocol(ipPermission types.IpPermission) string {
	protocol, ok := protocols[*ipPermission.IpProtocol]
	if !ok {
		protocol = "Unknown"
	}

	return protocol
}

func tabulateIpPermissions(ipPermissions []types.IpPermission, securityGroup types.SecurityGroup, header string) string {
	ipPermissionsTable := table.NewWriter()

	ipPermissionsTable.SetTitle(header)
	ipPermissionsTable.AppendHeader(table.Row{"Protocol", "Port Range", "Source", "Description"})
	ipPermissionsTable.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:      1,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignDefault,
		},
		{
			Number:      2,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignRight,
			AutoMerge:   true,
		},
		{
			Number:      3,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignDefault,
		},
		{
			Number:      4,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignDefault,
		},
	})
	ipPermissionsTable.Style().Box = table.StyleBoxRounded
	ipPermissionsTable.Style().Format = table.FormatOptions{
		Header: text.FormatDefault,
	}
	ipPermissionsTable.Style().Title.Align = text.AlignCenter

	for _, ipPermission := range ipPermissions {
		protocol := determineProtocol(ipPermission)
		portRange := determinePortRange(ipPermission)

		for _, ipRange := range ipPermission.IpRanges {
			source := *ipRange.CidrIp
			description := ""
			if ipRange.Description != nil {
				description = *ipRange.Description
			}

			ipPermissionsTable.AppendRow(table.Row{
				protocol,
				portRange,
				source,
				description,
			})
		}

		for _, ipv6Range := range ipPermission.Ipv6Ranges {
			source := *ipv6Range.CidrIpv6
			description := ""
			if ipv6Range.Description != nil {
				description = *ipv6Range.Description
			}

			ipPermissionsTable.AppendRow(table.Row{
				protocol,
				portRange,
				source,
				description,
			})
		}

		for _, prefixListId := range ipPermission.PrefixListIds {
			source := *prefixListId.PrefixListId
			description := ""
			if prefixListId.Description != nil {
				description = *prefixListId.Description
			}

			ipPermissionsTable.AppendRow(table.Row{
				protocol,
				portRange,
				source,
				description,
			})
		}

		for _, userIdGroupPair := range ipPermission.UserIdGroupPairs {
			source := *userIdGroupPair.GroupId
			if userIdGroupPair.UserId != nil && *securityGroup.OwnerId != *userIdGroupPair.UserId {
				source = *userIdGroupPair.UserId + "/" + source
			}
			description := ""
			if userIdGroupPair.Description != nil {
				description = *userIdGroupPair.Description
			}

			ipPermissionsTable.AppendRow(table.Row{
				protocol,
				portRange,
				source,
				description,
			})
		}
	}

	return ipPermissionsTable.Render()
}

func tabulateSecurityGroup(securityGroup types.SecurityGroup) string {
	securityGroupTable := table.NewWriter()

	securityGroupTable.SetTitle(*securityGroup.GroupName)
	securityGroupTable.AppendHeader(table.Row{"VPC ID", "Group ID", "Group Name", "Description", "Owner"})
	securityGroupTable.SetColumnConfigs([]table.ColumnConfig{
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
		{
			Number:      3,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignCenter,
		},
		{
			Number:      4,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignCenter,
		},
		{
			Number:      5,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignCenter,
		},
	})
	securityGroupTable.Style().Box = table.StyleBoxRounded
	securityGroupTable.Style().Format = table.FormatOptions{
		Header: text.FormatDefault,
	}
	securityGroupTable.Style().Title.Align = text.AlignCenter

	securityGroupTable.AppendRow(table.Row{
		*securityGroup.VpcId,
		*securityGroup.GroupId,
		*securityGroup.GroupName,
		*securityGroup.Description,
		*securityGroup.OwnerId,
	})

	return securityGroupTable.Render()
}

func tabulateTags(tags []types.Tag, header string) string {
	tagsTable := table.NewWriter()

	tagsTable.SetTitle(header)
	tagsTable.AppendHeader(table.Row{"Key", "Value"})
	tagsTable.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:      1,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignLeft,
		},
		{
			Number:      2,
			AlignHeader: text.AlignCenter,
			Align:       text.AlignLeft,
		},
	})
	tagsTable.Style().Box = table.StyleBoxRounded
	tagsTable.Style().Format = table.FormatOptions{
		Header: text.FormatDefault,
	}
	tagsTable.Style().Title.Align = text.AlignCenter

	for _, tag := range tags {
		tagsTable.AppendRow(table.Row{
			*tag.Key,
			*tag.Value,
		})
	}

	return tagsTable.Render()
}