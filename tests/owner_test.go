/*
 *    Copyright 2026 InfAI (CC SES)
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func TestImportTypeOwner(t *testing.T) {
	log.InitForTest()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	permClient, conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	base := "http://localhost:" + conf.ServerPort + "/import-types/"
	admin, err := createTokenWithRoles("test", "admin-1", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	author, err := createToken("test", "author-1")
	if err != nil {
		t.Fatal(err)
	}
	move := func(token jwt.Token, id string, body string) (int, []byte) {
		resp, err := jwtput(token, base+url.PathEscape(id)+"/owner", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, b
	}
	read := func(id string) model.ImportType {
		resp, err := jwtget(admin, base+url.PathEscape(id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var result model.ImportType
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}

	b := new(bytes.Buffer)
	_ = json.NewEncoder(b).Encode(model.ImportType{Name: "owner-test", Description: "d", Cost: 3, Output: model.ContentVariable{Name: "out", FunctionId: "f1", AspectIds: []string{"a1"}}})
	resp, err := jwtpost(author, base, "application/json", b)
	if err != nil {
		t.Fatal(err)
	}
	var created model.ImportType
	err = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	if err != nil || created.Owner != "author-1" {
		t.Fatalf("created %+v, %v", created, err)
	}

	t.Run("not an admin", func(t *testing.T) {
		if status, body := move(author, created.Id, `{"owner":"content"}`); status != http.StatusForbidden {
			t.Errorf("status %d: %s", status, body)
		}
		if read(created.Id).Owner != "author-1" {
			t.Error("the owner changed")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"owner":" content"}`, `content`} {
			if status, resp := move(admin, created.Id, body); status != http.StatusBadRequest {
				t.Errorf("%q: status %d: %s", body, status, resp)
			}
		}
	})
	t.Run("unknown id", func(t *testing.T) {
		if status, body := move(admin, "urn:infai:ses:import-type:unknown", `{"owner":"content"}`); status != http.StatusNotFound {
			t.Errorf("status %d: %s", status, body)
		}
	})
	t.Run("get answers 405", func(t *testing.T) {
		resp, err := jwtget(admin, base+url.PathEscape(created.Id)+"/owner")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status %d", resp.StatusCode)
		}
	})
	t.Run("move", func(t *testing.T) {
		status, body := move(admin, created.Id, `{"owner":"content"}`)
		if status != http.StatusOK {
			t.Fatalf("status %d: %s", status, body)
		}
		var answer model.ImportType
		if err := json.Unmarshal(body, &answer); err != nil {
			t.Fatal(err)
		}
		want := created
		want.Owner = "content"
		if !reflect.DeepEqual(answer, want) || !reflect.DeepEqual(read(created.Id), want) {
			t.Errorf("answer %+v, stored %+v, want only the owner changed: %+v", answer, read(created.Id), want)
		}
		resource, err, _ := permClient.GetResource(permV2.InternalAdminToken, "import-types", created.Id)
		if err != nil {
			t.Fatal(err)
		}
		full := resource.UserPermissions["content"]
		if !(full.Read && full.Write && full.Execute && full.Administrate) || !resource.UserPermissions["author-1"].Read || !resource.RolePermissions["admin"].Administrate {
			t.Errorf("permissions %+v", resource.ResourcePermissions)
		}
	})
	t.Run("repeat", func(t *testing.T) {
		if status, body := move(admin, created.Id, `{"owner":"content"}`); status != http.StatusOK {
			t.Errorf("status %d: %s", status, body)
		}
	})
	t.Run("criteria filter still finds it", func(t *testing.T) {
		resp, err := jwtget(admin, base+"?criteria="+url.QueryEscape(`[{"function_id":"f1","aspect_ids":["a1"]}]`))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var list []model.ImportType
		if err = json.NewDecoder(resp.Body).Decode(&list); err != nil || len(list) != 1 || list[0].Id != created.Id {
			t.Errorf("list %+v, %v", list, err)
		}
	})
}
