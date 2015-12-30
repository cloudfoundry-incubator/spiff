package flow

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"
)

var embeddedDynaml = regexp.MustCompile(`^\(\((.*)\)\)$`)

func Flow(source yaml.Node, stubs ...yaml.Node) (yaml.Node, error) {
	result := source

	for {
		debug.Debug("@@@ loop:  %+v\n",result)
		next := flow(result, Environment{Stubs: stubs}, true)
		debug.Debug("@@@ --->   %+v\n",next)

		if reflect.DeepEqual(result, next) {
			break
		}

		result = next
	}
	debug.Debug("@@@ Done\n")
	unresolved := dynaml.FindUnresolvedNodes(result)
	if len(unresolved) > 0 {
		return nil, dynaml.UnresolvedNodes{unresolved}
	}

	return result, nil
}

func flow(root yaml.Node, env Environment, shouldOverride bool) yaml.Node {
	if root == nil {
		return root
	}
	
	replace:= root.ReplaceFlag()
	redirect:= root.RedirectPath()
	if redirect != nil {
		env = env.RedirectOverwrite(redirect)
	}
	
	if !replace {
		switch val := root.Value().(type) {
		case map[string]yaml.Node:
			return flowMap(root, env)

		case []yaml.Node:
			return flowList(root, env)

		case dynaml.Expression:
			debug.Debug("??? eval %+v\n", val)
			result, info, ok := val.Evaluate(env)
			if !ok {
				debug.Debug("??? failed ---> KEEP\n")
				return root
			}
			if info.Replace {
				debug.Debug("   REPLACE\n")
			}
			if len(info.RedirectPath) > 0 {
				redirect = info.RedirectPath
				debug.Debug("??? m--> %+v -> %v\n", result, info.RedirectPath)
				if !info.Replace {
					return yaml.RedirectNode(result.Value(), result, redirect)
				}
			}
			if info.Replace {
				result=yaml.ReplaceNode(result.Value(), result, redirect)
			}
			debug.Debug("??? ---> %+v\n", result)
			return result
	
		case string:
			result := flowString(root, env)
			if result != nil {
				_, ok := result.Value().(dynaml.Expression)
				if ok {
					return result
				}
			}
		}
	}

	if shouldOverride {
		debug.Debug("/// lookup %v -> %v\n",env.Path, env.StubPath)
		overridden, found := env.FindInStubs(env.StubPath)
		if found {
			root = overridden
		}
	}

    if replace {
		return yaml.ReplaceNode(root.Value(),root,redirect)
	}
	if redirect != nil {
		return yaml.RedirectNode(root.Value(),root,redirect)
	}

	return root
}

func optionalMerge(initial bool, node yaml.Node) bool {
	/////////////////////////////////////////////////////////////
	// compatibility check. A single merge node is always optional
	// means: <<: (( merge )) == <<: (( merge || nil ))
	// the first pass, just parses the dynaml
	// only the second pass, evaluates a dynaml node!
	if !initial {
		merge, ok := node.Value().(dynaml.MergeExpr)
		return ok && !merge.Required
	}
	return false
	/////////////////////////////////////////////////////////////
}

func flowMap(root yaml.Node, env Environment) yaml.Node {
	processed := true
	rootMap := root.Value().(map[string]yaml.Node)

	env = env.WithScope(rootMap)
	
	redirect:= root.RedirectPath()
    replace:= root.ReplaceFlag()
	newMap := make(map[string]yaml.Node)

	sortedKeys := getSortedKeys(rootMap)

	debug.Debug("HANDLE MAP %v\n", env.Path)
	// iteration order matters for the "<<" operator, it must be the first key in the map that is handled
	for i := range sortedKeys {
		key := sortedKeys[i]
		val := rootMap[key]

		if key == "<<" {
			_, initial := val.Value().(string)
			base := flow(val, env, false)
			_, ok := base.Value().(dynaml.Expression)
			if ok {
				if optionalMerge(initial,base) {
					continue
				}
				/////////////////////////////////////////////////////////////
				val = base
				processed = false;
			} else {
				baseMap, ok := base.Value().(map[string]yaml.Node)
				if base != nil && base.RedirectPath() != nil {
					redirect = base.RedirectPath()
					env=env.RedirectOverwrite(redirect)
				}
				if ok {
					for k, v := range baseMap {
						newMap[k] = v
					}
				}
				replace=base.ReplaceFlag()
				if replace {
				  break
				}
				continue
			}
		} else {
			if processed {
				val = flow(val, env.WithPath(key), true)
			}
		}

		newMap[key] = val
	}

	debug.Debug("MAP DONE %v\n", env.Path)
	if replace { 
	  return yaml.ReplaceNode(newMap, root, redirect)
	}
	return yaml.RedirectNode(newMap, root, redirect)
}

func flowList(root yaml.Node, env Environment) yaml.Node {
	rootList := root.Value().([]yaml.Node)

	debug.Debug("HANDLE LIST %v\n", env.Path)
	merged, process, replaced, redirectPath := processMerges(root, rootList, env)
	
	if process {

		newList := []yaml.Node{}
        if len(redirectPath) > 0 {
			env=env.RedirectOverwrite(redirectPath)
		}
		for idx, val := range merged {
			step := stepName(idx, val)
			newList = append(newList, flow(val, env.WithPath(step), false))
		}

		merged = newList
	}
	
	debug.Debug("LIST DONE %v\n", env.Path)
	if replaced {
		return yaml.ReplaceNode(merged, root, redirectPath)
	}
	if len(redirectPath) > 0 {
		return yaml.RedirectNode(merged, root, redirectPath)
	}
	return yaml.SubstituteNode(merged, root)
}

func flowString(root yaml.Node, env Environment) yaml.Node {
	rootString := root.Value().(string)

	sub := embeddedDynaml.FindStringSubmatch(rootString)
	if sub == nil {
		return root
	}
	debug.Debug("dynaml: %v: %s\n", env.Path, sub[1])
	expr, err := dynaml.Parse(sub[1], env.Path)
	if err != nil {
		return root
	}

	return yaml.SubstituteNode(expr, root)
}

func stepName(index int, value yaml.Node) string {
	name, ok := yaml.FindString(value, "name")
	if ok {
		return name
	}

	return fmt.Sprintf("[%d]", index)
}

func processMerges(orig yaml.Node, root []yaml.Node, env Environment) ([]yaml.Node, bool, bool, []string) {
	spliced := []yaml.Node{}
	process := true
	replaced:= orig.ReplaceFlag()
	redirectPath := orig.RedirectPath()
	
	for _, val := range root {
		if val == nil {
			continue
		}

		inlineNode, ok := yaml.UnresolvedMerge(val)
		if ok {
			debug.Debug("*** %+v\n",inlineNode.Value())
			_, initial := inlineNode.Value().(string)
			result := flow(inlineNode, env, false)
			debug.Debug("=== %+v\n",result)
			_, ok := result.Value().(dynaml.Expression)
			if ok {
				if optionalMerge(initial,inlineNode) {
					continue
				}
				newMap := make(map[string]yaml.Node) 
				newMap["<<"] = result
				val = yaml.SubstituteNode(newMap,orig)
				process = false
			} else {
				inline, ok := result.Value().([]yaml.Node)

				if ok {
					inlineNew := newEntries(inline, root)
					replaced = result.ReplaceFlag()
					redirectPath = result.RedirectPath()
					if replaced {
						spliced = inlineNew
						process = false
						break
					} else {
					  spliced = append(spliced, inlineNew...)
					}
				}
				continue
			}
		}

		spliced = append(spliced, val)
	}

	debug.Debug("--> %+v  proc=%v replaced=%v redirect=%v\n",spliced, process, replaced, redirectPath)
	return spliced, process, replaced, redirectPath
}

func newEntries(a []yaml.Node, b []yaml.Node) []yaml.Node {
	added := []yaml.Node{}

	for _, val := range a {
		name, ok := yaml.FindString(val, "name")
		if ok {
			_, found := yaml.Find(yaml.NewNode(b, "some map"), name) // TODO
			if found {
				continue
			}
		}

		added = append(added, val)
	}

	return added
}

func getSortedKeys(unsortedMap map[string]yaml.Node) []string {
	keys := make([]string, len(unsortedMap))
	i := 0
	for k, _ := range unsortedMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
