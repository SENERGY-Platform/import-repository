/*
 * Copyright 2020 InfAI (CC SES)
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

package mongo

import (
	"context"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/import-repository/lib/testutils/docker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMigration(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	_, ip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://" + ip + ":27017"

	getColorFunction := "urn:infai:ses:measuring-function:getColorFunction"
	getHumidityFunction := "urn:infai:ses:measuring-function:getHumidityFunction"

	testCharacteristic := "urn:infai:ses:characteristic:test"

	deviceAspect := "urn:infai:ses:aspect:deviceAspect"
	airAspect := "urn:infai:ses:aspect:airAspect"

	it1 := model.ImportType{
		Id:   "it1",
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
	}

	it2 := model.ImportType{
		Id:   "it2",
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
	}

	it3 := model.ImportType{
		Id:   "it3",
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
	}

	itNone := model.ImportType{
		Id:   "none",
		Name: "none",
		Output: model.ContentVariable{
			Name: "output",
		},
	}

	t.Run("create import types without criteria", func(t *testing.T) {
		subCtx, cancel := context.WithCancel(ctx)
		defer time.Sleep(time.Second)
		defer cancel()
		defer time.Sleep(time.Second)
		db, err := New(conf, subCtx, wg)
		if err != nil {
			t.Error(err)
			return
		}
		for _, importType := range []model.ImportType{it1, it2, it3, itNone} {
			_, err := db.importTypeCollection().ReplaceOne(ctx, bson.M{idKey: importType.Id}, importType, options.Replace().SetUpsert(true))
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("migrate", func(t *testing.T) {
		subCtx, cancel := context.WithCancel(ctx)
		defer time.Sleep(time.Second)
		defer cancel()
		defer time.Sleep(time.Second)
		db, err := New(conf, subCtx, wg)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("check migrations", func(t *testing.T) {
			c, err := db.importTypeCollection().Find(ctx, bson.M{})
			if err != nil {
				t.Error(err)
				return
			}
			var list []ImportTypeWithCriteria
			err = c.All(ctx, &list)
			if err != nil {
				t.Error(err)
				return
			}
			expected := []ImportTypeWithCriteria{
				{
					ImportType: it1,
					Criteria:   contentVariableToCertList(it1.Output),
				},
				{
					ImportType: it2,
					Criteria:   contentVariableToCertList(it2.Output),
				},
				{
					ImportType: it3,
					Criteria:   contentVariableToCertList(it3.Output),
				},
				{
					ImportType: itNone,
					Criteria:   contentVariableToCertList(itNone.Output),
				},
			}
			slices.SortFunc(expected, func(a, b ImportTypeWithCriteria) int {
				return strings.Compare(a.ImportType.Id, b.ImportType.Id)
			})
			slices.SortFunc(list, func(a, b ImportTypeWithCriteria) int {
				return strings.Compare(a.ImportType.Id, b.ImportType.Id)
			})
			if !reflect.DeepEqual(list, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", list, expected)
			}
		})
	})

	t.Run("check migration rerun", func(t *testing.T) {
		subCtx, cancel := context.WithCancel(ctx)
		defer time.Sleep(time.Second)
		defer cancel()
		defer time.Sleep(time.Second)
		db, err := New(conf, subCtx, wg)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("check migrations", func(t *testing.T) {
			c, err := db.importTypeCollection().Find(ctx, bson.M{})
			if err != nil {
				t.Error(err)
				return
			}
			var list []ImportTypeWithCriteria
			err = c.All(ctx, &list)
			if err != nil {
				t.Error(err)
				return
			}
			expected := []ImportTypeWithCriteria{
				{
					ImportType: it1,
					Criteria:   contentVariableToCertList(it1.Output),
				},
				{
					ImportType: it2,
					Criteria:   contentVariableToCertList(it2.Output),
				},
				{
					ImportType: it3,
					Criteria:   contentVariableToCertList(it3.Output),
				},
				{
					ImportType: itNone,
					Criteria:   contentVariableToCertList(itNone.Output),
				},
			}
			slices.SortFunc(expected, func(a, b ImportTypeWithCriteria) int {
				return strings.Compare(a.ImportType.Id, b.ImportType.Id)
			})
			slices.SortFunc(list, func(a, b ImportTypeWithCriteria) int {
				return strings.Compare(a.ImportType.Id, b.ImportType.Id)
			})
			if !reflect.DeepEqual(list, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", list, expected)
			}
		})
	})
}
