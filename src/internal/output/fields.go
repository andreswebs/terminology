package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ValidateFields parses a comma-separated list of field paths and verifies each
// against the JSON paths available on envelope, returning the cleaned list or an
// ErrInvalidField error naming the offending path and the valid ones.
func ValidateFields(paths string, envelope any) ([]string, error) {
	paths = strings.TrimSpace(paths)
	if paths == "" {
		return nil, nil
	}

	fields := strings.Split(paths, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	valid := collectJSONPaths(reflect.TypeOf(envelope), "")

	for _, f := range fields {
		if !valid[f] {
			validList := sortedKeys(valid)
			return nil, fmt.Errorf("%w: %s; valid paths: %s",
				ErrInvalidField, f, strings.Join(validList, ", "))
		}
	}

	return fields, nil
}

// ProjectFields reduces the marshaled JSON in data to the requested field paths,
// always retaining schema_version and ok, and returns the re-marshaled result.
func ProjectFields(data []byte, fields []string) ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]any)

	if v, ok := raw["schema_version"]; ok {
		result["schema_version"] = v
	}
	if v, ok := raw["ok"]; ok {
		result["ok"] = v
	}

	trie := buildTrie(fields)

	for key, node := range trie {
		if v, ok := raw[key]; ok {
			if len(node.children) == 0 {
				result[key] = v
			} else {
				result[key] = projectValue(v, node)
			}
		}
	}

	return json.Marshal(result)
}

type trieNode struct {
	children map[string]*trieNode
}

func buildTrie(fields []string) map[string]*trieNode {
	root := make(map[string]*trieNode)
	for _, f := range fields {
		parts := strings.Split(f, ".")
		cur := root
		for _, p := range parts {
			if cur[p] == nil {
				cur[p] = &trieNode{children: make(map[string]*trieNode)}
			}
			node := cur[p]
			cur = node.children
		}
	}
	return root
}

func projectValue(v any, node *trieNode) any {
	if len(node.children) == 0 {
		return v
	}

	switch val := v.(type) {
	case map[string]any:
		return projectMap(val, node.children)
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			if m, ok := item.(map[string]any); ok {
				result[i] = projectMap(m, node.children)
			} else {
				result[i] = item
			}
		}
		return result
	default:
		return v
	}
}

func projectMap(m map[string]any, children map[string]*trieNode) map[string]any {
	result := make(map[string]any)

	if wild, ok := children["*"]; ok {
		for k, v := range m {
			if len(wild.children) == 0 {
				result[k] = v
			} else {
				result[k] = projectValue(v, wild)
			}
		}
		return result
	}

	for key, node := range children {
		if v, ok := m[key]; ok {
			if len(node.children) == 0 {
				result[key] = v
			} else {
				result[key] = projectValue(v, node)
			}
		}
	}
	return result
}

func collectJSONPaths(t reflect.Type, prefix string) map[string]bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	paths := make(map[string]bool)

	if t.Kind() != reflect.Struct {
		return paths
	}

	for f := range t.Fields() {
		f := f
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			continue
		}

		fullPath := name
		if prefix != "" {
			fullPath = prefix + "." + name
		}

		paths[fullPath] = true

		ft := f.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Slice {
			ft = ft.Elem()
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
		}

		if ft.Kind() == reflect.Struct {
			for k := range collectJSONPaths(ft, fullPath) {
				paths[k] = true
			}
		}

		if ft.Kind() == reflect.Map {
			elemType := ft.Elem()
			if elemType.Kind() == reflect.Pointer {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				wildcardPrefix := fullPath + ".*"
				paths[wildcardPrefix] = true
				for k := range collectJSONPaths(elemType, wildcardPrefix) {
					paths[k] = true
				}
			}
		}
	}

	return paths
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
