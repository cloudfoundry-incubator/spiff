package dynaml

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var fileCache = map[string][]byte{}

func func_read(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) > 2 {
		info.Issue = yaml.NewIssue("read takes a maximum of two arguments")
		return nil, info, false
	}

	file, ok := arguments[0].(string)
	if !ok {
		info.Issue = yaml.NewIssue("string value requiredfor file path")
		return nil, info, false
	}

	t := "text"
	if strings.HasSuffix(file, ".yml") {
		t = "yaml"
	}
	if len(arguments) > 1 {
		t, ok = arguments[1].(string)
		if !ok {
			info.Issue = yaml.NewIssue("string value required for type")
			return nil, info, false
		}

	}

	var err error

	data := fileCache[file]
	if data == nil {
		debug.Debug("reading %s file %s\n", t, file)
		data, err = ioutil.ReadFile(file)
		if err != nil {
			info.Issue = yaml.NewIssue("error reading [%s]: %s", path.Clean(file), err)
			return nil, info, false
		}
		fileCache[file] = data
	}

	switch t {
	case "yaml":
		node, err := yaml.Parse(file, data)
		if err != nil {
			info.Issue = yaml.NewIssue("error parsing stub [%s]: %s", path.Clean(file), err)
			return nil, info, false
		}
		debug.Debug("resolving yaml file\n")
		result, state := binding.Flow(node, false)
		if state != nil {
			debug.Debug("resolving yaml file failed: " + state.Error())
			info.Issue = state.Issue("yaml file resolution failed")
			return nil, info, true
		}
		debug.Debug("resolving yaml file succeeded")
		info.Source = file
		return result.Value(), info, true

	case "text":
		info.Source = file
		return string(data), info, true

	default:
		info.Issue = yaml.NewIssue("invalid file type [%s] %s", path.Clean(file), t)
		return nil, info, false
	}

}
