package dynaml

import (
	"strings"
    "strconv"
	
	"github.com/cloudfoundry-incubator/spiff/yaml"
)


func func_join(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info:= DefaultInfo()
	
	if len(arguments)<1 {
		return nil,info,false
	}
	
	args := make([]string,0)
	for i, arg := range arguments {
		switch arg.(type) {
			case string:
				args=append(args,arg.(string))
			case int64:
				args=append(args,strconv.FormatInt(arg.(int64), 10))
			case []yaml.Node:
				if i==0 {
					return nil, info, false
				}
				elems, _ := arg.([]yaml.Node)
				for _, elem := range elems {
					switch elem.Value().(type) {
						case string:
							args=append(args,elem.Value().(string))
						case int64:
							args=append(args,strconv.FormatInt(elem.Value().(int64), 10))
						default:
							return nil, info, false
					}
				}
			default:
				return nil, info, false
		}
	}
		
	return node(strings.Join(args[1:],args[0])), info, true
}


