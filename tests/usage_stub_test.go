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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// usageStub plays the smart-service-repository and import-deploy: it answers the usage of an
// import type as set, zero for every other id.
type usageStub struct {
	mux      sync.Mutex
	releases map[string]int64
	imports  map[string]int64
	down     bool
	askedBy  []string
}

// currentUsage is the stub of the test environment created last; the tests run one after another.
var currentUsage *usageStub

func startUsageStub(ctx context.Context) (*usageStub, string) {
	stub := &usageStub{releases: map[string]int64{}, imports: map[string]int64{}}
	server := httptest.NewServer(stub)
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return stub, server.URL
}

func (s *usageStub) set(releases map[string]int64, imports map[string]int64, down bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.releases, s.imports, s.down, s.askedBy = releases, imports, down, nil
}

func (s *usageStub) tokens() []string {
	s.mux.Lock()
	defer s.mux.Unlock()
	return append([]string{}, s.askedBy...)
}

func (s *usageStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.askedBy = append(s.askedBy, r.Header.Get("Authorization"))
	if s.down {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if id, ok := strings.CutPrefix(r.URL.Path, "/resource-usage/import-types/"); ok {
		_ = json.NewEncoder(w).Encode(map[string]any{"releases": s.releases[id], "instances": s.releases[id] * 2, "readable": []any{}})
		return
	}
	if id, ok := strings.CutPrefix(r.URL.Path, "/import-type-usage/"); ok {
		_ = json.NewEncoder(w).Encode(map[string]any{"instances": s.imports[id], "readable": []any{}})
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
