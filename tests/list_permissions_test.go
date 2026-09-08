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
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

// TestListPermissions checks that the list is restricted to the import types the caller may
// read, both without an id filter and with one.
func TestListPermissions(t *testing.T) {
	log.InitForTest()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	user1, err := createToken("test", "user1")
	if err != nil {
		t.Error(err)
		return
	}
	user2, err := createToken("test", "user2")
	if err != nil {
		t.Error(err)
		return
	}
	admin, err := createTokenWithRoles("test", "admin", []string{"admin"})
	if err != nil {
		t.Error(err)
		return
	}

	c := client.NewClient("http://localhost:" + conf.ServerPort)

	create := func(token jwt.Token, name string) (result model.ImportType) {
		t.Helper()
		result, err, _ := c.CreateImportType(model.ImportType{Name: name, Output: model.ContentVariable{Name: "output"}}, token)
		if err != nil {
			t.Error(err)
		}
		return result
	}

	u1a := create(user1, "u1-a")
	u1b := create(user1, "u1-b")
	u2a := create(user2, "u2-a")
	if t.Failed() {
		return
	}

	t.Run("user1", testImportTypesListWithToken(c, user1, client.ImportTypeListOptions{}, []model.ImportType{u1a, u1b}))
	t.Run("user2", testImportTypesListWithToken(c, user2, client.ImportTypeListOptions{}, []model.ImportType{u2a}))
	t.Run("admin", testImportTypesListWithToken(c, admin, client.ImportTypeListOptions{}, []model.ImportType{u1a, u1b, u2a}))

	allIds := []string{u1a.Id, u1b.Id, u2a.Id}
	t.Run("user1 by ids", testImportTypesListWithToken(c, user1, client.ImportTypeListOptions{Ids: allIds}, []model.ImportType{u1a, u1b}))
	t.Run("user2 by ids", testImportTypesListWithToken(c, user2, client.ImportTypeListOptions{Ids: allIds}, []model.ImportType{u2a}))
}

func testImportTypesListWithToken(c client.Interface, token jwt.Token, options client.ImportTypeListOptions, expected []model.ImportType) func(t *testing.T) {
	return func(t *testing.T) {
		result, total, err, _ := c.ListImportTypes(token, options)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("\n%#v\n%#v\n", result, expected)
			return
		}
		if total != int64(len(expected)) {
			t.Errorf("total=%v, expected %v", total, len(expected))
		}
	}
}
