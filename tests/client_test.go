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
	"net/http"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/client"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
)

// The client sent create and update without a Content-Type, and the update as POST /import-types:
// every write stored an import type that had nothing but the id and owner the server sets.
func TestClientWritesTheImportType(t *testing.T) {
	log.InitForTest()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	user, err := createToken("test", "user1")
	if err != nil {
		t.Fatal(err)
	}
	c := client.NewClient("http://localhost:" + conf.ServerPort)

	created, err, _ := c.CreateImportType(model.ImportType{Name: "first", Output: model.ContentVariable{Name: "output"}}, user)
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "first" || created.Output.Name != "output" {
		t.Fatalf("create answered name %q, output name %q", created.Name, created.Output.Name)
	}
	stored, err, _ := c.ReadImportType(created.Id, user)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "first" || stored.Output.Name != "output" {
		t.Fatalf("create stored name %q, output name %q", stored.Name, stored.Output.Name)
	}

	stored.Name = "second"
	err, code := c.SetImportType(stored, user)
	if err != nil || code != http.StatusOK {
		t.Fatalf("set answered %d: %v", code, err)
	}
	updated, err, _ := c.ReadImportType(created.Id, user)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "second" || updated.Output.Name != "output" || updated.Owner != created.Owner {
		t.Errorf("set stored name %q, output name %q, owner %q", updated.Name, updated.Output.Name, updated.Owner)
	}
	_, total, err, _ := c.ListImportTypes(user, model.ImportTypeListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Errorf("%d import types after create and set, want 1", total)
	}

	unknown := stored
	unknown.Id = "urn:infai:ses:import-type:unknown"
	err, code = c.SetImportType(unknown, user)
	if err == nil {
		t.Errorf("set of an unknown id answered %d without an error", code)
	}
}
