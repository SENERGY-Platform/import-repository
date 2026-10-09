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
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func TestImportTypeDeleteChecksUsage(t *testing.T) {
	log.InitForTest()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	stub := currentUsage
	create := func(t *testing.T) string {
		it, err := createImportType(conf, model.ImportType{Name: "usage-test"})
		if err != nil {
			t.Fatal(err)
		}
		return it.Id
	}
	remove := func(token jwt.Token, id string, query string) (status int, body string) {
		resp, err := jwtdelete(token, "http://localhost:"+conf.ServerPort+"/import-types/"+url.PathEscape(id)+query)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	exists := func(id string) bool {
		resp, err := jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/import-types/"+url.PathEscape(id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}

	t.Run("used by releases", func(t *testing.T) {
		id := create(t)
		stub.set(map[string]int64{id: 2}, nil, false)
		status, body := remove(userjwt, id, "")
		if status != http.StatusConflict {
			t.Fatalf("status %d: %s", status, body)
		}
		var answer model.InUseResponse
		if err := json.Unmarshal([]byte(body), &answer); err != nil || answer.Releases != 2 || answer.ReleaseInstances != 4 {
			t.Errorf("body %s (%v)", body, err)
		}
		if !exists(id) {
			t.Error("the import type is gone")
		}
		for _, auth := range stub.tokens() {
			if auth != userjwt.Token {
				t.Errorf("usage asked with %q, want the caller's token", auth)
			}
		}
	})
	t.Run("used by import instances", func(t *testing.T) {
		id := create(t)
		stub.set(nil, map[string]int64{id: 3}, false)
		status, body := remove(userjwt, id, "")
		if status != http.StatusConflict {
			t.Fatalf("status %d: %s", status, body)
		}
		var answer model.InUseResponse
		if err := json.Unmarshal([]byte(body), &answer); err != nil || answer.ImportInstances != 3 {
			t.Errorf("body %s (%v)", body, err)
		}
		if !exists(id) {
			t.Error("the import type is gone")
		}
	})
	t.Run("unused", func(t *testing.T) {
		id := create(t)
		stub.set(nil, nil, false)
		if status, body := remove(userjwt, id, ""); status != http.StatusNoContent {
			t.Fatalf("status %d: %s", status, body)
		}
		if exists(id) {
			t.Error("the import type is still there")
		}
	})
	t.Run("a source is down", func(t *testing.T) {
		id := create(t)
		stub.set(nil, nil, true)
		for _, query := range []string{"", "?force=true"} {
			if status, body := remove(userjwt, id, query); status != http.StatusBadGateway {
				t.Errorf("%q: status %d: %s", query, status, body)
			}
		}
		if !exists(id) {
			t.Error("the import type is gone")
		}
	})
	t.Run("admin force", func(t *testing.T) {
		id := create(t)
		stub.set(map[string]int64{id: 1}, map[string]int64{id: 1}, false)
		if status, body := remove(userjwt, id, "?force=true"); status != http.StatusNoContent {
			t.Fatalf("status %d: %s", status, body)
		}
		if exists(id) {
			t.Error("the import type is still there")
		}
	})
	t.Run("force without the admin role", func(t *testing.T) {
		id := create(t)
		stub.set(map[string]int64{id: 1}, nil, false)
		user, err := createToken("test", "user1")
		if err != nil {
			t.Fatal(err)
		}
		if status, body := remove(user, id, "?force=true"); status != http.StatusForbidden {
			t.Errorf("status %d: %s", status, body)
		}
		if !exists(id) {
			t.Error("the import type is gone")
		}
	})
}
