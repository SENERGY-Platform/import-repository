/*
 * Copyright 2024 InfAI (CC SES)
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

package tests

import (
	"context"
	"github.com/SENERGY-Platform/import-repository/lib/client"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"reflect"
	"sync"
	"testing"
)

func TestList(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	getColorFunction := "urn:infai:ses:measuring-function:getColorFunction"
	getHumidityFunction := "urn:infai:ses:measuring-function:getHumidityFunction"

	testCharacteristic := "urn:infai:ses:characteristic:test"

	deviceAspect := "urn:infai:ses:aspect:deviceAspect"
	airAspect := "urn:infai:ses:aspect:airAspect"

	it1, err := createImportType(conf, model.ImportType{
		Name: "it1",
		Output: model.ContentVariable{
			Name: "output",
			SubContentVariables: []model.ContentVariable{
				{
					Name: "value",
					SubContentVariables: []model.ContentVariable{
						{
							Name:             "value",
							CharacteristicId: testCharacteristic,
							AspectId:         deviceAspect,
							FunctionId:       getColorFunction,
						},
						{
							Name:       "foo",
							AspectId:   airAspect,
							FunctionId: getHumidityFunction,
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	it2, err := createImportType(conf, model.ImportType{
		Name: "it2",
		Output: model.ContentVariable{
			Name: "output",
			SubContentVariables: []model.ContentVariable{
				{
					Name:             "value",
					CharacteristicId: testCharacteristic,
					AspectId:         deviceAspect,
					FunctionId:       getColorFunction,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	it3, err := createImportType(conf, model.ImportType{
		Name: "it3",
		Output: model.ContentVariable{
			Name: "output",
			SubContentVariables: []model.ContentVariable{
				{
					Name:       "foo",
					AspectId:   airAspect,
					FunctionId: getHumidityFunction,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	itNone, err := createImportType(conf, model.ImportType{
		Name: "none",
		Output: model.ContentVariable{
			Name: "output",
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	c := client.NewClient("http://localhost:" + conf.ServerPort)

	t.Run("default", testImportTypesList(c, client.ImportTypeListOptions{}, []model.ImportType{it1, it2, it3, itNone}))

	t.Run("sort name.asc", testImportTypesList(c, client.ImportTypeListOptions{SortBy: "name.asc"}, []model.ImportType{it1, it2, it3, itNone}))
	t.Run("sort name.desc", testImportTypesList(c, client.ImportTypeListOptions{SortBy: "name.desc"}, []model.ImportType{itNone, it3, it2, it1}))

	t.Run("limit offset", testImportTypesList(c, client.ImportTypeListOptions{Limit: 2, Offset: 1}, []model.ImportType{it2, it3}))

	t.Run("search it", testImportTypesList(c, client.ImportTypeListOptions{Search: "it"}, []model.ImportType{it1, it2, it3}))

	t.Run("search none", testImportTypesList(c, client.ImportTypeListOptions{Search: "none"}, []model.ImportType{itNone}))

	t.Run("criteria humidity device", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{deviceAspect}},
	}}, []model.ImportType{}))

	t.Run("criteria color device", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getColorFunction, AspectIds: []string{deviceAspect}},
	}}, []model.ImportType{it1, it2}))

	t.Run("criteria humidity air", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{airAspect}},
	}}, []model.ImportType{it1, it3}))

	t.Run("criteria humidity air,unknown", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{airAspect, "unknown"}},
	}}, []model.ImportType{it1, it3}))

	t.Run("multiple criteria", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{airAspect, "unknown"}},
		{FunctionId: getColorFunction, AspectIds: []string{deviceAspect}},
	}}, []model.ImportType{it1}))
}

func testImportTypesList(c client.Interface, options client.ImportTypeListOptions, expected []model.ImportType) func(t *testing.T) {
	return func(t *testing.T) {
		result, _, err, _ := c.ListImportTypes(userjwt, options)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("\n%#v\n%#v\n", result, expected)
			return
		}
	}
}
