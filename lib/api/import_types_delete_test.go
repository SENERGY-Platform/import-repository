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
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func init() {
	log.InitForTest()
	gin.SetMode(gin.TestMode)
}

// deleteController answers the delete as set up and records the options it got.
type deleteController struct {
	Controller
	err   error
	code  int
	asked *model.DeleteOptions
}

func (c *deleteController) DeleteImportTypeWithOptions(_ string, _ jwt.Token, opts model.DeleteOptions) (error, int) {
	c.asked = &opts
	return c.err, c.code
}

func unsignedToken(t *testing.T) string {
	t.Helper()
	encode := base64.RawURLEncoding.EncodeToString
	payload, err := json.Marshal(map[string]any{"sub": "alice"})
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + encode([]byte(`{"alg":"none","typ":"JWT"}`)) + "." + encode(payload) + "."
}

func deleteRequest(t *testing.T, control Controller, query string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.Use(gin_mw.ErrorHandler(model.GetStatusCode, ", "))
	ImportTypesEndpoints(config.Config{}, control, router)
	req := httptest.NewRequest(http.MethodDelete, "/import-types/urn:infai:ses:import-type:t1"+query, nil)
	req.Header.Set("Authorization", unsignedToken(t))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestDeleteAnswers204(t *testing.T) {
	control := &deleteController{code: http.StatusNoContent}
	recorder := deleteRequest(t, control, "")
	if recorder.Code != http.StatusNoContent {
		t.Errorf("answered %d: %s", recorder.Code, recorder.Body.String())
	}
	if control.asked == nil || control.asked.Force {
		t.Errorf("options = %+v, want no force", control.asked)
	}
}

func TestDeleteInUseAnswers409WithTheUsage(t *testing.T) {
	control := &deleteController{code: http.StatusConflict, err: &model.InUseError{Usage: model.ImportTypeUsage{
		SmartServices: model.SmartServiceUsage{Releases: 2, Instances: 5, Readable: []model.ReleaseRef{{Id: "r1", DesignId: "d1", Name: "heating"}}},
		Imports:       model.ImportUsage{Instances: 3},
	}}}
	recorder := deleteRequest(t, control, "")
	if recorder.Code != http.StatusConflict {
		t.Fatalf("answered %d: %s", recorder.Code, recorder.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not JSON: %v: %s", err, recorder.Body.String())
	}
	want := map[string]any{
		"error":             "import type is still used by smart services or import instances",
		"releases":          float64(2),
		"release_instances": float64(5),
		"import_instances":  float64(3),
		"readable_releases": []any{map[string]any{"id": "r1", "design_id": "d1", "name": "heating"}},
		"readable_imports":  []any{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}

func TestDeleteWhenUsageIsUnknownAnswers502(t *testing.T) {
	control := &deleteController{code: http.StatusBadGateway, err: model.ErrUsageUnavailable}
	recorder := deleteRequest(t, control, "?force=true")
	if recorder.Code != http.StatusBadGateway {
		t.Errorf("answered %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestDeleteForceIsPassedOn(t *testing.T) {
	for query, want := range map[string]bool{"?force=true": true, "?force=1": true, "?force=false": false, "?force=": false} {
		control := &deleteController{code: http.StatusNoContent}
		if recorder := deleteRequest(t, control, query); recorder.Code != http.StatusNoContent {
			t.Errorf("%s answered %d", query, recorder.Code)
		}
		if control.asked == nil || control.asked.Force != want {
			t.Errorf("%s: options = %+v, want force %v", query, control.asked, want)
		}
	}
}

func TestDeleteForceForNonAdminAnswers403(t *testing.T) {
	control := &deleteController{code: http.StatusForbidden, err: errors.New("force needs the admin role")}
	recorder := deleteRequest(t, control, "?force=true")
	if recorder.Code != http.StatusForbidden {
		t.Errorf("answered %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestDeleteForceMustBeABoolean(t *testing.T) {
	control := &deleteController{code: http.StatusNoContent}
	recorder := deleteRequest(t, control, "?force=maybe")
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("answered %d: %s", recorder.Code, recorder.Body.String())
	}
	if control.asked != nil {
		t.Errorf("the controller was called with %+v", control.asked)
	}
}
