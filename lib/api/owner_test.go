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

package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

type ownerController struct {
	Controller
	result model.ImportType
	err    error
	code   int
	called bool
	id     string
	owner  string
}

func (c *ownerController) SetImportTypeOwner(id string, owner string, _ jwt.Token) (model.ImportType, error, int) {
	c.called, c.id, c.owner = true, id, owner
	return c.result, c.err, c.code
}

func adminToken(t *testing.T) string {
	t.Helper()
	encode := base64.RawURLEncoding.EncodeToString
	payload, err := json.Marshal(map[string]any{"sub": "admin-user", "realm_access": map[string][]string{"roles": {"admin"}}})
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + encode([]byte(`{"alg":"none","typ":"JWT"}`)) + "." + encode(payload) + "."
}

func ownerRequest(t *testing.T, control Controller, method string, body string) *httptest.ResponseRecorder {
	t.Helper()
	return ownerRequestWith(t, control, method, body, adminToken(t))
}

func ownerRequestWith(t *testing.T, control Controller, method string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()
	router := newRouter(config.Config{}, control)
	req := httptest.NewRequest(method, "/import-types/it-1/owner", strings.NewReader(body))
	req.Header.Set("Authorization", token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestOwnerAnswersTheMovedImportType(t *testing.T) {
	control := &ownerController{result: model.ImportType{Id: "it-1", Name: "n", Owner: "user-c"}, code: http.StatusOK}
	recorder := ownerRequest(t, control, http.MethodPut, `{"owner": "user-c"}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("answered %d: %s", recorder.Code, recorder.Body.String())
	}
	if control.id != "it-1" || control.owner != "user-c" {
		t.Errorf("controller got id %q, owner %q", control.id, control.owner)
	}
	if !strings.Contains(recorder.Body.String(), `"owner":"user-c"`) || !strings.Contains(recorder.Body.String(), `"id":"it-1"`) {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestOwnerPassesAMissingOwnerOnAsEmpty(t *testing.T) {
	control := &ownerController{err: errors.New("owner missing"), code: http.StatusBadRequest}
	recorder := ownerRequest(t, control, http.MethodPut, `{}`)
	if !control.called || control.owner != "" || recorder.Code != http.StatusBadRequest {
		t.Errorf("called %v with owner %q, answered %d", control.called, control.owner, recorder.Code)
	}
}

func TestOwnerRejectsABodyThatIsNoJson(t *testing.T) {
	control := &ownerController{code: http.StatusOK}
	recorder := ownerRequest(t, control, http.MethodPut, `user-c`)
	if recorder.Code != http.StatusBadRequest || control.called {
		t.Errorf("answered %d, controller called %v, want 400 and no call", recorder.Code, control.called)
	}
}

func TestOwnerRefusesANonAdminBeforeReadingTheBody(t *testing.T) {
	control := &ownerController{code: http.StatusOK}
	recorder := ownerRequestWith(t, control, http.MethodPut, `user-c`, unsignedToken(t))
	if recorder.Code != http.StatusForbidden || control.called {
		t.Errorf("answered %d, controller called %v, want 403 and no call", recorder.Code, control.called)
	}
}

func TestOwnerAnswersRefusalsWithTheirStatus(t *testing.T) {
	for _, code := range []int{http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
		control := &ownerController{err: errors.New("refused"), code: code}
		recorder := ownerRequest(t, control, http.MethodPut, `{"owner": "user-c"}`)
		if recorder.Code != code {
			t.Errorf("answered %d, want %d", recorder.Code, code)
		}
	}
}

// The transfer tool probes the route with a GET and tells "route exists" from "old version" by 405.
func TestGetOnTheOwnerPathAnswers405(t *testing.T) {
	control := &ownerController{}
	recorder := ownerRequest(t, control, http.MethodGet, "")
	if recorder.Code != http.StatusMethodNotAllowed || control.called {
		t.Errorf("answered %d, controller called %v, want 405", recorder.Code, control.called)
	}
}
