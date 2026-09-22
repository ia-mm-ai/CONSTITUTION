package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGenesisSchemaMatchesRuntimeInventoryAndFixedBindings(t *testing.T) {
	raw, err := os.ReadFile("genesis/genesis.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	genesisBytes, _, _ := testGenesis(t)
	var genesis any
	if err := json.Unmarshal(genesisBytes, &genesis); err != nil {
		t.Fatal(err)
	}
	var check func(string, map[string]any, any)
	check = func(path string, node map[string]any, value any) {
		if reference, ok := node["$ref"].(string); ok {
			name := strings.TrimPrefix(reference, "#/$defs/")
			check(path, schema["$defs"].(map[string]any)[name].(map[string]any), value)
			return
		}
		if constant, fixed := node["const"]; fixed {
			if !reflect.DeepEqual(constant, value) {
				t.Fatalf("schema binding drift at %s", path)
			}
			return
		}
		switch value := value.(type) {
		case map[string]any:
			properties, ok := node["properties"].(map[string]any)
			if !ok || node["additionalProperties"] != false {
				t.Fatalf("schema object %s must explicitly classify all fields", path)
			}
			required := node["required"].([]any)
			if len(required) != len(value) || len(properties) != len(value) {
				t.Fatalf("schema field inventory drift at %s", path)
			}
			for _, name := range required {
				if _, exists := value[name.(string)]; !exists {
					t.Fatalf("schema requires unknown field %s.%s", path, name)
				}
			}
			for name, child := range value {
				property, exists := properties[name]
				if !exists {
					t.Fatalf("schema omitted field %s.%s", path, name)
				}
				check(path+"."+name, property.(map[string]any), child)
			}
		case []any:
			options := node["items"].(map[string]any)["enum"].([]any)
			if node["uniqueItems"] != true || node["minItems"] != float64(len(value)) || node["maxItems"] != float64(len(value)) || len(options) != len(value) {
				t.Fatalf("schema set cardinality drift at %s", path)
			}
			for _, item := range value {
				found := false
				for _, option := range options {
					found = found || item == option
				}
				if !found {
					t.Fatalf("schema set member drift at %s", path)
				}
			}
		}
	}
	check("genesis", schema, genesis)
}
