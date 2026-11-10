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

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

type writeController struct {
	Controller
	got    model.ImportType
	called bool
}

func (c *writeController) CreateImportType(importType model.ImportType, _ jwt.Token) (model.ImportType, error, int) {
	c.called, c.got = true, importType
	return importType, nil, http.StatusOK
}

func (c *writeController) SetImportType(importType model.ImportType, _ jwt.Token) (error, int) {
	c.called, c.got = true, importType
	return nil, http.StatusOK
}

func writeRequest(t *testing.T, control Controller, method string, path string, contentType string, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := newRouter(config.Config{}, control)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", adminToken(t))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

// Without a Content-Type, gin's ShouldBind picks form binding and drops a JSON body without an error,
// so the import type was stored with nothing but the id and owner the server sets.
func TestWritesBindTheJsonBodyWithoutContentType(t *testing.T) {
	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/import-types", `{"name":"n","output":{"name":"out"}}`},
		{http.MethodPut, "/import-types/it-1", `{"id":"it-1","name":"n","output":{"name":"out"}}`},
	}
	for _, tc := range cases {
		for _, contentType := range []string{"", "application/json"} {
			control := &writeController{}
			recorder := writeRequest(t, control, tc.method, tc.path, contentType, tc.body)
			if recorder.Code != http.StatusOK {
				t.Errorf("%s with Content-Type %q answered %d: %s", tc.method, contentType, recorder.Code, recorder.Body.String())
				continue
			}
			if control.got.Name != "n" || control.got.Output.Name != "out" {
				t.Errorf("%s with Content-Type %q passed name %q, output name %q", tc.method, contentType, control.got.Name, control.got.Output.Name)
			}
		}
	}
}

func TestWritesRejectAFormBody(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/import-types"},
		{http.MethodPut, "/import-types/it-1"},
	}
	for _, tc := range cases {
		control := &writeController{}
		recorder := writeRequest(t, control, tc.method, tc.path, "application/x-www-form-urlencoded", "id=it-1&name=n")
		if recorder.Code != http.StatusBadRequest || control.called {
			t.Errorf("%s answered %d, controller called %v, want 400 and no call", tc.method, recorder.Code, control.called)
		}
	}
}
