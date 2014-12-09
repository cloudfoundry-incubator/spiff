package dynaml

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	switch e.Name {
	case "static_ips":
		indices := make([]int, len(e.Arguments))
		for i, arg := range e.Arguments {
			index, ok := arg.Evaluate(binding)
			if !ok {
				return nil, false
			}

			index64, ok := index.Value().(int64)
			if !ok {
				return nil, false
			}
			indices[i] = int(index64)
		}

		return generateStaticIPs(binding, indices)
	}

	return nil, false
}

func (e CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, e := range e.Arguments {
		args[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
}

func generateStaticIPs(binding Binding, indices []int) (yaml.Node, bool) {
	if len(indices) == 0 {
		return nil, false
	}

	ranges, ok := findStaticIPRanges(binding)
	if !ok {
		return nil, false
	}

	instanceCount, ok := findInstanceCount(binding)
	if !ok {
		return nil, false
	}

	ipPool, ok := staticIPPool(ranges)
	if !ok {
		return nil, false
	}

	ips := []yaml.Node{}
	for _, i := range indices {
		if len(ipPool) <= i {
			return nil, false
		}

		ips = append(ips, node(ipPool[i].String()))
	}

	if len(ips) < instanceCount {
		return nil, false
	}

	return node(ips[:instanceCount]), true
}

func findInstanceCount(binding Binding) (int, bool) {
	nearestInstances, found := binding.FindReference([]string{"instances"})
	if !found {
		return 0, false
	}

	instances, ok := nearestInstances.Value().(int64)
	return int(instances), ok
}

func findStaticIPRanges(binding Binding) ([]string, bool) {
	nearestNetworkName, found := binding.FindReference([]string{"name"})
	if !found {
		return nil, false
	}

	networkName, ok := nearestNetworkName.Value().(string)
	if !ok {
		return nil, false
	}

	subnets, found := binding.FindFromRoot(
		[]string{"networks", networkName, "subnets"},
	)

	if !found {
		return nil, false
	}

	subnetsList, ok := subnets.Value().([]yaml.Node)
	if !ok {
		return nil, false
	}

	allRanges := []string{}

	for _, subnet := range subnetsList {
		subnetMap, ok := subnet.Value().(map[string]yaml.Node)
		if !ok {
			return nil, false
		}

		static, ok := subnetMap["static"]

		if !ok {
			return nil, false
		}

		staticList, ok := static.Value().([]yaml.Node)
		if !ok {
			return nil, false
		}

		ranges := make([]string, len(staticList))

		for i, r := range staticList {
			ipsString, ok := r.Value().(string)
			if !ok {
				return nil, false
			}

			ranges[i] = ipsString
		}

		allRanges = append(allRanges, ranges...)
	}

	return allRanges, true
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

		start, end = sortIpRangeBoundaries(start, end)
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

func sortIpRangeBoundaries(a, b net.IP) (net.IP, net.IP) {
	cmp := bytes.Compare(a, b)
	if cmp > 0 {
		return b, a
	} else if cmp < 0 {
		return a, b
	}

	return a, b
}
