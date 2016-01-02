package dynaml

import (
	"fmt"
	"log"
	"strings"
    "strconv"
	"os/exec"
	"crypto/md5"
	
	"github.com/cloudfoundry-incubator/candiedyaml"
	
	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"
)


func func_exec(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info:= DefaultInfo()
	
	if len(arguments)<1 {
		return nil,info,false
	}
	args := []string{}
	debug.Debug("exec: found %d arguments for call\n",len(arguments))
	for i, arg := range arguments {
		list, ok := arg.([]yaml.Node)
		if i==0 && ok {
			debug.Debug("exec: found array as first argument\n")
			if len(arguments)==1 && len(list)>0 {
				// handle single list argument to gain command and argument
				for j,arg := range list {
					v, ok := getArg(j,arg.Value())
					if !ok {
						info.Issue="command argument must be string"
						return nil, info, false
					}
					args =append(args,v)
				}
			} else {
				info.Issue="list not allowed for command argument"
				return nil, info, false
			}
		} else {
			v, ok := getArg(i,arg)
			if !ok {
				info.Issue="command argument must be string"
				return nil, info, false
			}
			args = append(args,v)
		}
	}
	result,err := cachedExecute(args)
	if err!=nil {
		info.Issue="execution '"+args[0]+"' failed"
		// expression set to undefined
		return nil, info, false
	}
		
	str:=string(result)
	execYML,err := yaml.Parse("exec",result)
	if strings.HasPrefix(str, "---\n") && err==nil {
		debug.Debug("exec: found yaml result %+v\n",execYML)
		return execYML, info, true
	} else {
		if strings.HasSuffix(str,"\n") {
			str=str[:len(str)-1]
		}
		int64YML,err := strconv.ParseInt(str,10,64)
		if err == nil {
			debug.Debug("exec: found integer result: %s\n",int64YML)
			return node(int64YML), info, true
		}
		debug.Debug("exec: found string result: %s\n",string(result))
		return node(str), info, true
	}
}

func getArg(i int, value interface{}) (string, bool) {
	debug.Debug("arg %d: %+v\n", i, value)
	switch value.(type) {
		case string:
			return value.(string), true
		case int64:
			return strconv.FormatInt(value.(int64), 10), true
		default:
			if i==0 {
				return "", false
			}
			yaml, err := candiedyaml.Marshal(node(value))
			if err != nil {
				log.Fatalln("error marshalling manifest:", err)
			}
			return "---\n"+string(yaml), true
	}
}

var cache = make(map[string][]byte)

func cachedExecute(args []string) ([]byte, error) {
	h := md5.New()
	for _, arg := range args {
		h.Write([]byte(arg))
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))
	result := cache[hash]
	if result != nil {
		debug.Debug("exec: reusing cache %s for %v\n",hash, args)
		return result, nil
	}
	debug.Debug("exec: calling %v\n",args)
	cmd := exec.Command(args[0], args[1:]...)
	result,err := cmd.Output()
	cache[hash]=result
	return result, err
}

