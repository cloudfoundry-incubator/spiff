package dynaml

import (
	"net"
	"strings"

	"github.com/vito/spiff/yaml"
)

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) Evaluate(context Context) (yaml.Node, bool) {
	switch e.Name {
	case "static_ips":
		indices := make([]int, len(e.Arguments))
		for i, arg := range e.Arguments {
			index, ok := arg.Evaluate(context)
			if !ok {
				return nil, false
			}

			indices[i], ok = index.(int)
			if !ok {
				return nil, false
			}
		}

		return generateStaticIPs(context, indices)
	default:
		return nil, false
	}
}

func generateStaticIPs(context Context, indices []int) (yaml.Node, bool) {
	if len(indices) == 0 {
		return nil, false
	}

	ranges, ok := findStaticIPRanges(context)
	if !ok {
		return nil, false
	}

	instanceCount, ok := findInstanceCount(context)
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

		ips = append(ips, ipPool[i].String())
	}

	if len(ips) < instanceCount {
		return nil, false
	}

	return ips[:instanceCount], true
}

func findInstanceCount(context Context) (int, bool) {
	nearestInstances, found := context.FindReference([]string{"instances"})
	if !found {
		return 0, false
	}

	instances, ok := nearestInstances.(int)
	return instances, ok
}

func findStaticIPRanges(context Context) ([]string, bool) {
	nearestNetworkName, found := context.FindReference([]string{"name"})
	if !found {
		return nil, false
	}

	networkName, ok := nearestNetworkName.(string)
	if !ok {
		return nil, false
	}

	static, found := context.FindFromRoot(
		[]string{"networks", networkName, "subnets", "[0]", "static"},
	)

	if !found {
		return nil, false
	}

	staticList, ok := static.([]yaml.Node)
	if !ok {
		return nil, false
	}

	ranges := make([]string, len(staticList))

	for i, r := range staticList {
		ipsString, ok := r.(string)
		if !ok {
			return nil, false
		}

		ranges[i] = ipsString
	}

	return ranges, true
}

func staticIPPool(ranges []string) ([]net.IP, bool) {
	ipPool := []net.IP{}

	for _, r := range ranges {
		segments := strings.Split(r, "-")
		if len(segments) != 2 {
			return nil, false
		}

		start := net.ParseIP(strings.Trim(segments[0], " "))
		end := net.ParseIP(strings.Trim(segments[1], " "))

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
