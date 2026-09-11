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

package tests

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/client"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func TestContentVariableAspectIds(t *testing.T) {
	log.InitForTest()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, conf, deviceRepoDb, err := createTestEnvWithDeviceRepo(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	getHumidityFunction := "urn:infai:ses:measuring-function:getHumidityFunction"

	parentAspect := "urn:infai:ses:aspect:parent"
	childAspect1 := "urn:infai:ses:aspect:child1"
	childAspect2 := "urn:infai:ses:aspect:child2"
	unknownAspect := "urn:infai:ses:aspect:unknown"

	//the aspect hierarchy the AND filter resolves the subtrees from
	for _, node := range []models.AspectNode{
		{Id: parentAspect, RootId: parentAspect, ChildIds: []string{childAspect1, childAspect2}, DescendentIds: []string{childAspect1, childAspect2}},
		{Id: childAspect1, RootId: parentAspect, ParentId: parentAspect, AncestorIds: []string{parentAspect}},
		{Id: childAspect2, RootId: parentAspect, ParentId: parentAspect, AncestorIds: []string{parentAspect}},
	} {
		err = deviceRepoDb.SetAspectNode(ctx, node)
		if err != nil {
			t.Error(err)
			return
		}
	}

	//written the way a client that only knows the deprecated aspect_id writes it
	itDeprecated, err := createImportType(conf, model.ImportType{
		Name: "a-deprecated",
		Output: model.ContentVariable{
			Name: "output",
			SubContentVariables: []model.ContentVariable{
				{
					Name:       "value",
					AspectId:   childAspect1,
					FunctionId: getHumidityFunction,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	//AspectIds unsorted on purpose: the deprecated alias is the alphabetically first entry
	itBoth, err := createImportType(conf, model.ImportType{
		Name: "b-both",
		Output: model.ContentVariable{
			Name: "output",
			SubContentVariables: []model.ContentVariable{
				{
					Name:       "value",
					AspectIds:  []string{childAspect2, childAspect1},
					FunctionId: getHumidityFunction,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	itParent, err := createImportType(conf, model.ImportType{
		Name: "c-parent",
		Output: model.ContentVariable{
			Name: "output",
			SubContentVariables: []model.ContentVariable{
				{
					Name:       "value",
					AspectIds:  []string{parentAspect},
					FunctionId: getHumidityFunction,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	c := client.NewClient("http://localhost:" + conf.ServerPort)

	t.Run("aspect_id is an alias for a single element aspect_ids", func(t *testing.T) {
		result, err, _ := c.ReadImportType(itDeprecated.Id, userjwt)
		if err != nil {
			t.Error(err)
			return
		}
		variable := result.Output.SubContentVariables[0]
		if variable.AspectId != childAspect1 || !reflect.DeepEqual(variable.AspectIds, []string{childAspect1}) {
			t.Errorf("\n%#v\n", variable)
		}
	})

	t.Run("aspect_id is the alphabetically first of aspect_ids", func(t *testing.T) {
		result, err, _ := c.ReadImportType(itBoth.Id, userjwt)
		if err != nil {
			t.Error(err)
			return
		}
		variable := result.Output.SubContentVariables[0]
		if variable.AspectId != childAspect1 || !reflect.DeepEqual(variable.AspectIds, []string{childAspect2, childAspect1}) {
			t.Errorf("\n%#v\n", variable)
		}
	})

	//the default: the aspect_ids of a criteria are alternatives and cover no descendents
	t.Run("or child1", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{childAspect1}},
	}}, []model.ImportType{itDeprecated, itBoth}))

	t.Run("or child1,child2", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{childAspect1, childAspect2}},
	}}, []model.ImportType{itDeprecated, itBoth}))

	t.Run("or parent", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectIds: []string{parentAspect}},
	}}, []model.ImportType{itParent}))

	//a criteria that names only the deprecated aspect_id filters like a single element aspect_ids
	t.Run("or with only the deprecated aspect_id", testImportTypesList(c, client.ImportTypeListOptions{Criteria: []model.ImportTypeFilterCriteria{
		{FunctionId: getHumidityFunction, AspectId: childAspect1},
	}}, []model.ImportType{itDeprecated, itBoth}))

	t.Run("and child1", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectIds: []string{childAspect1}},
		}}, []model.ImportType{itDeprecated, itBoth}))

	//only a content variable that carries both aspects matches
	t.Run("and child1,child2", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectIds: []string{childAspect1, childAspect2}},
		}}, []model.ImportType{itBoth}))

	//an aspect of an AND criteria covers its subtree, which is resolved from the device-repository
	t.Run("and parent", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectIds: []string{parentAspect}},
		}}, []model.ImportType{itDeprecated, itBoth, itParent}))

	//the parent condition is met by a variable carrying a child, the child condition is not met
	//by a variable carrying only the parent
	t.Run("and parent,child1", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectIds: []string{parentAspect, childAspect1}},
		}}, []model.ImportType{itDeprecated, itBoth}))

	t.Run("and with only the deprecated aspect_id", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectId: childAspect1},
		}}, []model.ImportType{itDeprecated, itBoth}))

	//the deprecated aspect_id joins the list instead of replacing it, so both are demanded
	t.Run("and with the deprecated aspect_id next to the list", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectId: childAspect2, AspectIds: []string{childAspect1}},
		}}, []model.ImportType{itBoth}))

	//an unknown aspect covers only itself and is carried by nobody
	t.Run("and child1,unknown", testImportTypesList(c, client.ImportTypeListOptions{
		AndCombineCriteriaAspectIds: true,
		Criteria: []model.ImportTypeFilterCriteria{
			{FunctionId: getHumidityFunction, AspectIds: []string{childAspect1, unknownAspect}},
		}}, []model.ImportType{}))
}
