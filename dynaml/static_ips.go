package dynaml

import (
	"net"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var (
	refName      = ReferenceExpr{[]string{"name"}}
	refInstances = ReferenceExpr{[]string{"instances"}}
)

func func_static_ips(arguments []Expression, binding Binding) (yaml.Node, EvaluationInfo, bool) {

	indices := make([]int, len(arguments))
	for i, arg := range arguments {
		index, info, ok := arg.Evaluate(binding)
		if !ok {
			return nil, info, false
		}

		index64, ok := index.Value().(int64)
		if !ok {
			return nil, info, false
		}
		indices[i] = int(index64)
	}

	return generateStaticIPs(binding, indices)
}

func generateStaticIPs(binding Binding, indices []int) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(indices) == 0 {
		return nil, info, false
	}

	ranges, info, ok := findStaticIPRanges(binding)
	if !ok || ranges == nil {
		return nil, info, ok
	}

	instanceCountP, info, ok := findInstanceCount(binding)
	if !ok || instanceCountP == nil {
		return nil, info, ok
	}
	instanceCount := int(*instanceCountP)
	ipPool, ok := staticIPPool(ranges)
	if !ok {
		return nil, info, false
	}

	ips := []yaml.Node{}
	for _, i := range indices {
		if len(ipPool) <= i {
			return nil, info, false
		}

		ips = append(ips, node(ipPool[i].String()))
	}

	if len(ips) < instanceCount {
		info.Issue = yaml.NewIssue("too less static IPs for %d instances", instanceCount)
		return nil, info, false
	}

	return node(ips[:instanceCount]), info, true
}

func findInstanceCount(binding Binding) (*int64, EvaluationInfo, bool) {
	nearestInstances, info, found := refInstances.Evaluate(binding)
	if !found || isExpression(nearestInstances) {
		return nil, info, false
	}

	instances, ok := nearestInstances.Value().(int64)
	return &instances, info, ok
}

func findStaticIPRanges(binding Binding) ([]string, EvaluationInfo, bool) {
	nearestNetworkName, info, found := refName.Evaluate(binding)
	if !found || isExpression(nearestNetworkName) {
		return nil, info, found
	}

	networkName, ok := nearestNetworkName.Value().(string)
	if !ok {
		info.Issue = yaml.NewIssue("name field must be string")
		return nil, info, false
	}

	subnetsRef := ReferenceExpr{[]string{"", "networks", networkName, "subnets"}}
	subnets, info, found := subnetsRef.Evaluate(binding)

	if !found {
		return nil, info, false
	}
	if isExpression(subnets) {
		return nil, info, true
	}

	subnetsList, ok := subnets.Value().([]yaml.Node)
	if !ok {
		info.Issue = yaml.NewIssue("subnets field must be a list")
		return nil, info, false
	}

	allRanges := []string{}

	for _, subnet := range subnetsList {
		subnetMap, ok := subnet.Value().(map[string]yaml.Node)
		if !ok {
			info.Issue = yaml.NewIssue("subnet must be a map")
			return nil, info, false
		}

		static, ok := subnetMap["static"]

		if !ok {
			info.Issue = yaml.NewIssue("no static ips for network %s", networkName)
			return nil, info, false
		}

		staticList, ok := static.Value().([]yaml.Node)
		if !ok {
			info.Issue = yaml.NewIssue("static ips for network %s must be a list", networkName)
			return nil, info, false
		}

		ranges := make([]string, len(staticList))

		for i, r := range staticList {
			ipsString, ok := r.Value().(string)
			if !ok {
				info.Issue = yaml.NewIssue("invalid entry for static ips for network %s", networkName)
				return nil, info, false
			}

			ranges[i] = ipsString
		}

		allRanges = append(allRanges, ranges...)
	}

	return allRanges, info, true
}

func staticIPPool(ranges []string) ([]net.IP, bool) {
	ipPool := []net.IP{}

	for _, r := range ranges {
		segments := strings.Split(r, "-")
		if len(segments) == 0 {
			return nil, false
		}

		var start, end net.IP

		start = net.ParseIP(strings.Trim(segments[0], " "))

		if len(segments) == 1 {
			end = start
		} else {
			end = net.ParseIP(strings.Trim(segments[1], " "))
		}

		ipPool = append(ipPool, ipRange(start, end)...)
	}

	return ipPool, true
}

func ipRange(a, b net.IP) []net.IP {
	prev := a

	ips := []net.IP{a}

	for !prev.Equal(b) {
		next := net.ParseIP(prev.String())
		inc(next)
		ips = append(ips, next)
		prev = next
	}

	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
