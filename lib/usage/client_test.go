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

package usage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/model"
)

const importTypeId = "urn:infai:ses:import-type:t1"

// remote answers every request with the given status and body and records the last request.
func remote(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()
	last := &http.Request{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*last = *r.Clone(context.Background())
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, last
}

func TestSmartServices(t *testing.T) {
	server, last := remote(t, http.StatusOK, `{"releases":2,"instances":5,"readable":[{"id":"r1","design_id":"d1","name":"heating"}]}`)
	client := New(server.URL+"/", "")
	usage, err := client.SmartServices(context.Background(), importTypeId, "Bearer abc")
	if err != nil {
		t.Fatal(err)
	}
	want := model.SmartServiceUsage{Releases: 2, Instances: 5, Readable: []model.ReleaseRef{{Id: "r1", DesignId: "d1", Name: "heating"}}}
	if !reflect.DeepEqual(usage, want) {
		t.Errorf("usage = %+v, want %+v", usage, want)
	}
	if last.URL.Path != "/resource-usage/import-types/"+importTypeId {
		t.Errorf("path = %q", last.URL.Path)
	}
	if last.Header.Get("Authorization") != "Bearer abc" {
		t.Errorf("Authorization = %q, want the caller's token", last.Header.Get("Authorization"))
	}
}

func TestImports(t *testing.T) {
	server, last := remote(t, http.StatusOK, `{"instances":3,"readable":[{"id":"i1","name":"weather"}]}`)
	client := New("", server.URL)
	usage, err := client.Imports(context.Background(), importTypeId, "Bearer abc")
	if err != nil {
		t.Fatal(err)
	}
	want := model.ImportUsage{Instances: 3, Readable: []model.ImportRef{{Id: "i1", Name: "weather"}}}
	if !reflect.DeepEqual(usage, want) {
		t.Errorf("usage = %+v, want %+v", usage, want)
	}
	if last.URL.Path != "/import-type-usage/"+importTypeId {
		t.Errorf("path = %q", last.URL.Path)
	}
	if last.Header.Get("Authorization") != "Bearer abc" {
		t.Errorf("Authorization = %q, want the caller's token", last.Header.Get("Authorization"))
	}
}

func TestZeroCountsAreAnswersNotErrors(t *testing.T) {
	server, _ := remote(t, http.StatusOK, `{"releases":0,"instances":0,"readable":null}`)
	client := New(server.URL, server.URL)
	smart, err := client.SmartServices(context.Background(), importTypeId, "")
	if err != nil || smart.Releases != 0 || smart.Readable == nil {
		t.Errorf("smart services: %+v, %v", smart, err)
	}
	imports, err := client.Imports(context.Background(), importTypeId, "")
	if err != nil || imports.Instances != 0 || imports.Readable == nil {
		t.Errorf("imports: %+v, %v", imports, err)
	}
}

// An answer that is not the expected one must never read as "unused".
func TestUnknownAnswersAreErrors(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
	}{
		"401":                      {http.StatusUnauthorized, `unauthorized`},
		"403":                      {http.StatusForbidden, `{"releases":0,"instances":0}`},
		"404":                      {http.StatusNotFound, ``},
		"500":                      {http.StatusInternalServerError, `{"releases":0,"instances":0}`},
		"200 with an empty object": {http.StatusOK, `{}`},
		"200 with another form":    {http.StatusOK, `[{"id":"x"}]`},
		"200 with no body":         {http.StatusOK, ``},
		"200 with text":            {http.StatusOK, `ok`},
		"200 with a string count":  {http.StatusOK, `{"releases":"2","instances":"5"}`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			server, _ := remote(t, c.status, c.body)
			client := New(server.URL, server.URL)
			if usage, err := client.SmartServices(context.Background(), importTypeId, ""); err == nil {
				t.Errorf("smart services answered %+v", usage)
			}
			if usage, err := client.Imports(context.Background(), importTypeId, ""); err == nil {
				t.Errorf("imports answered %+v", usage)
			}
		})
	}
}

func TestSmartServicesNeedBothCounts(t *testing.T) {
	for _, body := range []string{`{"releases":1}`, `{"instances":1}`} {
		server, _ := remote(t, http.StatusOK, body)
		if usage, err := New(server.URL, "").SmartServices(context.Background(), importTypeId, ""); err == nil {
			t.Errorf("%s answered %+v", body, usage)
		}
	}
}

func TestUnreachableSourceIsAnError(t *testing.T) {
	server, _ := remote(t, http.StatusOK, `{}`)
	address := server.URL
	server.Close()
	client := New(address, address)
	if _, err := client.SmartServices(context.Background(), importTypeId, ""); err == nil {
		t.Error("smart services: expected an error")
	}
	if _, err := client.Imports(context.Background(), importTypeId, ""); err == nil {
		t.Error("imports: expected an error")
	}
}
