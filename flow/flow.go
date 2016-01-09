package flow

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func Flow(source yaml.Node, stubs ...yaml.Node) (yaml.Node, error) {
	result := source

	for {
		debug.Debug("@@@ loop:  %+v\n", result)
		next := flow(result, Environment{Stubs: stubs}, true)
		debug.Debug("@@@ --->   %+v\n", next)

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

	replace := root.ReplaceFlag()
	redirect := root.RedirectPath()
	preferred := root.Preferred()
	merged := root.Merged()
	keyName := root.KeyName()

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
			debug.Debug("??? eval %T: %+v\n", val, val)
			result, info, ok := val.Evaluate(env)
			if !ok {
				root = yaml.IssueNode(root, info.Issue)
				debug.Debug("??? failed ---> KEEP\n")
				if !shouldOverride {
					return root
				}
				replace = replace || info.Replace
			} else {
				_, ok = result.Value().(string)
				if ok {
					// map result to potential expression
					result = flowString(result, env)
				}
				_, expr := result.Value().(dynaml.Expression)

				// preserve accumulated node attributes
				if preferred || info.Preferred {
					debug.Debug("   PREFERRED")
					result = yaml.PreferredNode(result)
				}

				if info.KeyName != "" {
					keyName = info.KeyName
					result = yaml.KeyNameNode(result, keyName)
				}
				if len(info.RedirectPath) > 0 {
					redirect = info.RedirectPath
				}
				if len(redirect) > 0 {
					debug.Debug("   REDIRECT -> %v\n", redirect)
					result = yaml.RedirectNode(result.Value(), result, redirect)
				}

				if replace || info.Replace {
					debug.Debug("   REPLACE\n")
					result = yaml.ReplaceNode(result.Value(), result, redirect)
				} else {
					if merged || info.Merged {
						debug.Debug("   MERGED\n")
						result = yaml.MergedNode(result)
					}
				}
				if expr || result.Merged() || !shouldOverride || result.Preferred() {
					debug.Debug("   prefer expression over override")
					debug.Debug("??? ---> %+v\n", result)
					return result
				}
				debug.Debug("???   try override")
				replace = result.ReplaceFlag()
				root = result
			}

		case string:
			result := flowString(root, env)
			if result != nil {
				_, ok := result.Value().(dynaml.Expression)
				if ok {
					// analyse expression before overriding
					return result
				}
			}
		}
	}

	if !merged && shouldOverride {
		debug.Debug("/// lookup stub %v -> %v\n", env.Path, env.StubPath)
		overridden, found := env.FindInStubs(env.StubPath)
		if found {
			root = overridden
			if keyName != "" {
				root = yaml.KeyNameNode(root, keyName)
			}
			if replace {
				return yaml.ReplaceNode(root.Value(), root, redirect)
			}
			if redirect != nil {
				return yaml.RedirectNode(root.Value(), root, redirect)
			}
			if merged {
				return yaml.MergedNode(root)
			}
		}
	}

	return root
}

/*
 * compatibility issue. A single merge node was always optional
 * means: <<: (( merge )) == <<: (( merge || nil ))
 * the first pass, just parses the dynaml
 * only the second pass, evaluates a dynaml node!
 */
func simpleMergeCompatibilityCheck(initial bool, node yaml.Node) bool {
	if !initial {
		merge, ok := node.Value().(dynaml.MergeExpr)
		return ok && !merge.Required
	}
	return false
}

func flowMap(root yaml.Node, env Environment) yaml.Node {
	processed := true
	rootMap := root.Value().(map[string]yaml.Node)

	env = env.WithScope(rootMap)

	redirect := root.RedirectPath()
	replace := root.ReplaceFlag()
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
				if simpleMergeCompatibilityCheck(initial, base) {
					continue
				}
				val = base
				processed = false
			} else {
				baseMap, ok := base.Value().(map[string]yaml.Node)
				if base != nil && base.RedirectPath() != nil {
					redirect = base.RedirectPath()
					env = env.RedirectOverwrite(redirect)
				}
				if ok {
					for k, v := range baseMap {
						newMap[k] = v
					}
				}
				replace = base.ReplaceFlag()
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

		debug.Debug("MAP (%s)%s\n", val.KeyName(), key)
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
	merged, process, replaced, redirectPath, keyName := processMerges(root, rootList, env)

	if process {

		newList := []yaml.Node{}
		if len(redirectPath) > 0 {
			env = env.RedirectOverwrite(redirectPath)
		}
		for idx, val := range merged {
			step := stepName(idx, val, keyName)
			debug.Debug("  step %s\n", step)
			newList = append(newList, flow(val, env.WithPath(step), false))
		}

		merged = newList
	}

	if keyName != "" {
		root = yaml.KeyNameNode(root, keyName)
	}
	debug.Debug("LIST DONE (%s)%v\n", root.KeyName(), env.Path)
	if replaced {
		return yaml.ReplaceNode(merged, root, redirectPath)
	}
	if len(redirectPath) > 0 {
		return yaml.RedirectNode(merged, root, redirectPath)
	}
	return yaml.SubstituteNode(merged, root)
}

func flowString(root yaml.Node, env Environment) yaml.Node {

	sub := yaml.EmbeddedDynaml(root)
	if sub == nil {
		return root
	}
	debug.Debug("dynaml: %v: %s\n", env.Path, *sub)
	expr, err := dynaml.Parse(*sub, env.Path, env.StubPath)
	if err != nil {
		return root
	}

	return yaml.SubstituteNode(expr, root)
}

func stepName(index int, value yaml.Node, keyName string) string {
	if keyName == "" {
		keyName = "name"
	}
	name, ok := yaml.FindString(value, keyName)
	if ok {
		return keyName + ":" + name
	}

	return fmt.Sprintf("[%d]", index)
}

func processMerges(orig yaml.Node, root []yaml.Node, env Environment) ([]yaml.Node, bool, bool, []string, string) {
	spliced := []yaml.Node{}
	process := true
	keyName := orig.KeyName()
	replaced := orig.ReplaceFlag()
	redirectPath := orig.RedirectPath()

	for _, val := range root {
		if val == nil {
			continue
		}

		inlineNode, ok := yaml.UnresolvedListEntryMerge(val)
		if ok {
			debug.Debug("*** %+v\n", inlineNode.Value())
			_, initial := inlineNode.Value().(string)
			result := flow(inlineNode, env, false)
			if result.KeyName() != "" {
				keyName = result.KeyName()
			}
			debug.Debug("=== (%s)%+v\n", keyName, result)
			_, ok := result.Value().(dynaml.Expression)
			if ok {
				if simpleMergeCompatibilityCheck(initial, inlineNode) {
					continue
				}
				newMap := make(map[string]yaml.Node)
				newMap["<<"] = result
				val = yaml.SubstituteNode(newMap, orig)
				process = false
			} else {
				inline, ok := result.Value().([]yaml.Node)

				if ok {
					inlineNew := newEntries(inline, root, keyName)
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

		val, newKey := ProcessKeyTag(val)
		if newKey != "" {
			keyName = newKey
		}
		spliced = append(spliced, val)
	}

	debug.Debug("--> %+v  proc=%v replaced=%v redirect=%v key=%s\n", spliced, process, replaced, redirectPath, keyName)
	return spliced, process, replaced, redirectPath, keyName
}

func ProcessKeyTag(val yaml.Node) (yaml.Node, string) {
	keyName := ""

	m, ok := val.Value().(map[string]yaml.Node)
	if ok {
		found := false
		for key, _ := range m {
			split := strings.Index(key, ":")
			if split > 0 {
				if key[:split] == "key" {
					keyName = key[split+1:]
					found = true
				}
			}
		}
		if found {
			newMap := make(map[string]yaml.Node)
			for key, v := range m {
				split := strings.Index(key, ":")
				if split > 0 {
					if key[:split] == "key" {
						key = key[split+1:]
					}
				}
				newMap[key] = v
			}
			return yaml.SubstituteNode(newMap, val), keyName
		}
	}
	return val, keyName
}

func newEntries(a []yaml.Node, b []yaml.Node, keyName string) []yaml.Node {
	if keyName == "" {
		keyName = "name"
	}
	old := yaml.KeyNameNode(yaml.NewNode(b, "some map"), keyName)
	added := []yaml.Node{}

	for _, val := range a {
		name, ok := yaml.FindStringR(true, val, keyName)
		if ok {
			_, found := yaml.FindR(true, old, name) // TODO
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
