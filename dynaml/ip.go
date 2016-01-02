package dynaml

import (
	"fmt"
	"net"
	
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_ip(op func(ip net.IP, cidr *net.IPNet) net.IP, arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info:= DefaultInfo()
	
	if len(arguments)!=1 {
		info.Issue="only one argument expected for CIDR function"
		return nil,info,false
	}
	
	str, ok:= arguments[0].(string)
	if !ok {
		info.Issue="CIDR argument required"
		return nil, info, false
	}
	
	ip, cidr, err:=net.ParseCIDR(str)
	
	if err!=nil {
		info.Issue="CIDR argument required"
		return nil, info, false
	}
	
	return node(op(ip,cidr).String()), info, true
}

func func_minIP(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	return func_ip(func(ip net.IP, cidr *net.IPNet) net.IP {
		return ip.Mask(cidr.Mask)
	},arguments,binding)
}

func func_maxIP(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	return func_ip(func(ip net.IP, cidr *net.IPNet) net.IP {
		return MaxIP(cidr)
	},arguments,binding)
}

func SubIP(ip net.IP, mask net.IPMask) net.IP {
	m:=ip.Mask(mask)
	fmt.Printf("%d\n", len(m))
	out := make(net.IP, len(ip))
	for i, v:= range ip {
		j:=len(ip)-i
		if j>len(m) {
			out[i]=v
		} else {
		 	out[i]=v&^m[len(m)-j]
		}
	}
	return out
}

func MaxIP(cidr *net.IPNet) net.IP {
	ip:=cidr.IP.Mask(cidr.Mask)
	mask:=cidr.Mask
	out := make(net.IP, len(ip))
	for i, v:= range ip {
		j:=len(ip)-i
		if j>len(mask) {
			out[i]=v
		} else {
		 	out[i]=v|^mask[len(mask)-j]
		}
	}
	return out
}