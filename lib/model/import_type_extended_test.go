/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// ImportTypeExtended embeds the import type inline, so its own fields and the ones it inherits
// sit in the same document. A missing bson:",inline" still compiles and still passes every
// test that reads the struct in Go, while the stored form silently grows an importtype sub
// document; json needs no tag, it promotes an anonymous field on its own. These lists are the
// encoded form as it was before the embedding, and the json one is the contract of the kafka
// command this type is the payload of.

var importTypeExtendedJsonPaths = []string{
	"aspect_functions[]",
	"configs[].default_value",
	"configs[].description",
	"configs[].name",
	"configs[].type",
	"content_aspect_ids[]",
	"content_function_ids[]",
	"cost",
	"default_restart",
	"description",
	"id",
	"image",
	"name",
	"output.aspect_id",
	"output.aspect_ids[]",
	"output.characteristic_id",
	"output.function_id",
	"output.name",
	"output.sub_content_variables[].aspect_id",
	"output.sub_content_variables[].aspect_ids[]",
	"output.sub_content_variables[].characteristic_id",
	"output.sub_content_variables[].function_id",
	"output.sub_content_variables[].name",
	"output.sub_content_variables[].sub_content_variables",
	"output.sub_content_variables[].type",
	"output.sub_content_variables[].use_as_tag",
	"output.type",
	"output.use_as_tag",
	"owner",
}

// the bson names are the lowercased go field names, because these types carry no bson tags.
// defaultvaluestring is stored while it is absent from json, which is what it is for.
var importTypeExtendedBsonPaths = []string{
	"aspectfunctions[]",
	"configs[].defaultvalue",
	"configs[].defaultvaluestring",
	"configs[].description",
	"configs[].name",
	"configs[].type",
	"contentaspectids[]",
	"contentfunctionids[]",
	"cost",
	"defaultrestart",
	"description",
	"id",
	"image",
	"name",
	"output.aspectid",
	"output.aspectids[]",
	"output.characteristicid",
	"output.functionid",
	"output.name",
	"output.subcontentvariables[].aspectid",
	"output.subcontentvariables[].aspectids[]",
	"output.subcontentvariables[].characteristicid",
	"output.subcontentvariables[].functionid",
	"output.subcontentvariables[].name",
	"output.subcontentvariables[].subcontentvariables",
	"output.subcontentvariables[].type",
	"output.subcontentvariables[].useastag",
	"output.type",
	"output.useastag",
	"owner",
}

func TestImportTypeExtendedEncodedPaths(t *testing.T) {
	sample := fullImportTypeExtended()

	t.Run("the json document stays flat", func(t *testing.T) {
		b, err := json.Marshal(sample)
		if err != nil {
			t.Fatal(err)
		}
		decoded := map[string]interface{}{}
		if err = json.Unmarshal(b, &decoded); err != nil {
			t.Fatal(err)
		}
		assertPaths(t, encodedPaths("", decoded), importTypeExtendedJsonPaths)
	})

	t.Run("the bson document stays flat", func(t *testing.T) {
		b, err := bson.Marshal(sample)
		if err != nil {
			t.Fatal(err)
		}
		decoded := bson.M{}
		if err = bson.Unmarshal(b, &decoded); err != nil {
			t.Fatal(err)
		}
		assertPaths(t, encodedPaths("", decoded), importTypeExtendedBsonPaths)
	})
}

func assertPaths(t *testing.T, actual []string, expected []string) {
	t.Helper()
	if reflect.DeepEqual(actual, expected) {
		return
	}
	for _, path := range actual {
		if !slices.Contains(expected, path) {
			t.Errorf("unexpected path %v", path)
		}
	}
	for _, path := range expected {
		if !slices.Contains(actual, path) {
			t.Errorf("missing path %v", path)
		}
	}
}

// encodedPaths lists the field paths of a decoded document, an array marked as [] and its
// elements merged, so that the result describes the shape rather than the sample.
func encodedPaths(prefix string, value interface{}) []string {
	result := []string{}
	add := func(paths []string) {
		for _, path := range paths {
			if !slices.Contains(result, path) {
				result = append(result, path)
			}
		}
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, sub := range typed {
			add(encodedPaths(joinPath(prefix, key), sub))
		}
	case bson.M:
		for key, sub := range typed {
			add(encodedPaths(joinPath(prefix, key), sub))
		}
	case []interface{}:
		for _, sub := range typed {
			add(encodedPaths(prefix+"[]", sub))
		}
	case bson.A:
		for _, sub := range typed {
			add(encodedPaths(prefix+"[]", sub))
		}
	default:
		add([]string{prefix})
	}
	slices.Sort(result)
	return result
}

func joinPath(prefix string, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func fullImportTypeExtended() ImportTypeExtended {
	defaultValueString := "structured"
	return ImportTypeExtended{
		ImportType: ImportType{
			Id:             "id-1",
			Name:           "name-1",
			Description:    "desc-1",
			Image:          "image-1",
			DefaultRestart: true,
			Configs: []ImportConfig{{
				Name:               "config-1",
				Description:        "config-desc",
				Type:               Structure,
				DefaultValue:       "dv",
				DefaultValueString: &defaultValueString,
			}},
			Output: ContentVariable{
				Name:             "output",
				Type:             Structure,
				CharacteristicId: "char-0",
				SubContentVariables: []ContentVariable{{
					Name:             "value",
					Type:             String,
					CharacteristicId: "char-1",
					UseAsTag:         true,
					FunctionId:       "function-1",
					AspectId:         "aspect-1",
					AspectIds:        []string{"aspect-1", "aspect-2"},
				}},
				UseAsTag:   true,
				FunctionId: "function-0",
				AspectId:   "aspect-0",
				AspectIds:  []string{"aspect-0"},
			},
			Owner: "owner-1",
			Cost:  42,
		},
		ContentAspectIds:   []string{"aspect-1", "aspect-2"},
		ContentFunctionIds: []string{"function-1"},
		AspectFunctions:    []string{"aspect-1_function-1"},
	}
}
