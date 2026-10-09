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

package controller

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/database"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	permV2Model "github.com/SENERGY-Platform/permissions-v2/pkg/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

const importTypeId = "urn:infai:ses:import-type:t1"

func init() {
	log.InitForTest()
}

// removals records what the delete reached; the embedded interfaces stay nil because a delete
// must touch nothing else.
type removals struct {
	database.Database
	permV2.Client
	dbRemoved   []string
	permRemoved []string
	allowed     bool
}

func (r *removals) CheckPermission(_ string, _ string, _ string, _ ...permV2Model.Permission) (bool, error, int) {
	return r.allowed, nil, http.StatusOK
}

func (r *removals) RemoveResource(_ string, _ string, id string) (error, int) {
	r.permRemoved = append(r.permRemoved, id)
	return nil, http.StatusOK
}

func (r *removals) RemoveImportType(_ context.Context, id string) error {
	r.dbRemoved = append(r.dbRemoved, id)
	return nil
}

// usageAnswers returns canned answers and records the token it was asked with.
type usageAnswers struct {
	smartServices    model.SmartServiceUsage
	smartServicesErr error
	imports          model.ImportUsage
	importsErr       error
	askedWith        []string
}

func (u *usageAnswers) SmartServices(_ context.Context, _ string, auth string) (model.SmartServiceUsage, error) {
	u.askedWith = append(u.askedWith, auth)
	return u.smartServices, u.smartServicesErr
}

func (u *usageAnswers) Imports(_ context.Context, _ string, auth string) (model.ImportUsage, error) {
	u.askedWith = append(u.askedWith, auth)
	return u.imports, u.importsErr
}

func newDeleteController(usage UsageSource) (*Controller, *removals) {
	r := &removals{allowed: true}
	return &Controller{db: r, permV2Client: r, usage: usage}, r
}

func userToken() jwt.Token {
	return jwt.Token{Token: "Bearer user-token", Sub: "alice"}
}

func adminToken() jwt.Token {
	return jwt.Token{Token: "Bearer admin-token", Sub: "root", RealmAccess: map[string][]string{"roles": {"admin"}}}
}

func assertNothingDeleted(t *testing.T, r *removals) {
	t.Helper()
	if len(r.dbRemoved) > 0 || len(r.permRemoved) > 0 {
		t.Errorf("deleted although it must not: db %v, permissions %v", r.dbRemoved, r.permRemoved)
	}
}

func assertDeleted(t *testing.T, r *removals) {
	t.Helper()
	if len(r.dbRemoved) != 1 || r.dbRemoved[0] != importTypeId || len(r.permRemoved) != 1 || r.permRemoved[0] != importTypeId {
		t.Errorf("expected the import type removed once from both, got db %v, permissions %v", r.dbRemoved, r.permRemoved)
	}
}

func TestDeleteUnusedImportType(t *testing.T) {
	usage := &usageAnswers{}
	ctrl, r := newDeleteController(usage)
	err, code := ctrl.DeleteImportType(importTypeId, userToken())
	if err != nil || code != http.StatusNoContent {
		t.Fatalf("got %v, %d", err, code)
	}
	assertDeleted(t, r)
	for _, auth := range usage.askedWith {
		if auth != "Bearer user-token" {
			t.Errorf("usage asked with %q, want the caller's token", auth)
		}
	}
	if len(usage.askedWith) != 2 {
		t.Errorf("asked %d sources, want 2", len(usage.askedWith))
	}
}

func TestDeleteImportTypeUsedByReleases(t *testing.T) {
	usage := &usageAnswers{smartServices: model.SmartServiceUsage{
		Releases: 2, Instances: 5, Readable: []model.ReleaseRef{{Id: "r1", DesignId: "d1", Name: "heating"}},
	}}
	ctrl, r := newDeleteController(usage)
	err, code := ctrl.DeleteImportType(importTypeId, userToken())
	var inUse *model.InUseError
	if !errors.As(err, &inUse) || code != http.StatusConflict {
		t.Fatalf("got %v, %d, want an in-use error and 409", err, code)
	}
	if inUse.Usage.SmartServices.Releases != 2 || inUse.Usage.SmartServices.Instances != 5 || len(inUse.Usage.SmartServices.Readable) != 1 {
		t.Errorf("usage = %+v", inUse.Usage)
	}
	assertNothingDeleted(t, r)
}

func TestDeleteImportTypeUsedByImportInstances(t *testing.T) {
	usage := &usageAnswers{imports: model.ImportUsage{Instances: 3, Readable: []model.ImportRef{{Id: "i1", Name: "weather"}}}}
	ctrl, r := newDeleteController(usage)
	err, code := ctrl.DeleteImportType(importTypeId, userToken())
	var inUse *model.InUseError
	if !errors.As(err, &inUse) || code != http.StatusConflict {
		t.Fatalf("got %v, %d, want an in-use error and 409", err, code)
	}
	if inUse.Usage.Imports.Instances != 3 || len(inUse.Usage.Imports.Readable) != 1 {
		t.Errorf("usage = %+v", inUse.Usage)
	}
	assertNothingDeleted(t, r)
}

// Without the permission to delete, nobody learns who uses the import type.
func TestDeleteImportTypeWithoutRightsAsksNobody(t *testing.T) {
	usage := &usageAnswers{}
	ctrl, r := newDeleteController(usage)
	r.allowed = false
	err, code := ctrl.DeleteImportType(importTypeId, userToken())
	if err == nil || code != http.StatusForbidden {
		t.Fatalf("got %v, %d, want 403", err, code)
	}
	if len(usage.askedWith) != 0 {
		t.Errorf("usage was asked without the right to delete")
	}
	assertNothingDeleted(t, r)
}

func TestDeleteImportTypeWhenASourceIsDown(t *testing.T) {
	down := errors.New("connection refused: 10.0.0.1")
	cases := map[string]*usageAnswers{
		"smart service repository": {smartServicesErr: down},
		"import-deploy":            {importsErr: down},
	}
	for name, usage := range cases {
		for _, force := range []bool{false, true} {
			t.Run(name+" force="+map[bool]string{false: "false", true: "true"}[force], func(t *testing.T) {
				ctrl, r := newDeleteController(usage)
				err, code := ctrl.DeleteImportTypeWithOptions(importTypeId, adminToken(), model.DeleteOptions{Force: force})
				if !errors.Is(err, model.ErrUsageUnavailable) || code != http.StatusBadGateway {
					t.Fatalf("got %v, %d, want ErrUsageUnavailable and 502", err, code)
				}
				if strings.Contains(err.Error(), "10.0.0.1") {
					t.Errorf("the error shown to the client carries internals: %v", err)
				}
				assertNothingDeleted(t, r)
			})
		}
	}
}

func TestDeleteUsedImportTypeForcedByAdmin(t *testing.T) {
	var logged bytes.Buffer
	previous := log.Logger
	log.Logger = slog.New(slog.NewTextHandler(&logged, nil))
	t.Cleanup(func() { log.Logger = previous })

	usage := &usageAnswers{
		smartServices: model.SmartServiceUsage{Releases: 2, Instances: 5},
		imports:       model.ImportUsage{Instances: 3},
	}
	ctrl, r := newDeleteController(usage)
	err, code := ctrl.DeleteImportTypeWithOptions(importTypeId, adminToken(), model.DeleteOptions{Force: true})
	if err != nil || code != http.StatusNoContent {
		t.Fatalf("got %v, %d", err, code)
	}
	assertDeleted(t, r)
	line := logged.String()
	for _, want := range []string{"level=WARN", "import_type_id=" + importTypeId, "user_id=root", "releases=2", "release_instances=5", "import_instances=3"} {
		if !strings.Contains(line, want) {
			t.Errorf("log %q lacks %q", line, want)
		}
	}
}

// A forced delete of an unused import type is an ordinary delete and needs no warning.
func TestDeleteUnusedImportTypeForcedLogsNothing(t *testing.T) {
	var logged bytes.Buffer
	previous := log.Logger
	log.Logger = slog.New(slog.NewTextHandler(&logged, nil))
	t.Cleanup(func() { log.Logger = previous })

	ctrl, r := newDeleteController(&usageAnswers{})
	if err, code := ctrl.DeleteImportTypeWithOptions(importTypeId, adminToken(), model.DeleteOptions{Force: true}); err != nil || code != http.StatusNoContent {
		t.Fatalf("got %v, %d", err, code)
	}
	assertDeleted(t, r)
	if logged.Len() != 0 {
		t.Errorf("logged %q", logged.String())
	}
}

func TestDeleteImportTypeForceNeedsAdmin(t *testing.T) {
	usage := &usageAnswers{smartServices: model.SmartServiceUsage{Releases: 1}}
	ctrl, r := newDeleteController(usage)
	err, code := ctrl.DeleteImportTypeWithOptions(importTypeId, userToken(), model.DeleteOptions{Force: true})
	if err == nil || code != http.StatusForbidden {
		t.Fatalf("got %v, %d, want 403", err, code)
	}
	if len(usage.askedWith) != 0 {
		t.Errorf("usage was asked for a request that is refused anyway")
	}
	assertNothingDeleted(t, r)
}

func TestNewRefusesEmptyUsageAddresses(t *testing.T) {
	cases := map[string]config.Config{
		"no smart service repository": {ImportDeployUrl: "http://deploy"},
		"no import-deploy":            {SmartServiceRepoUrl: "http://smart-services"},
		"neither":                     {},
	}
	for name, conf := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl, err := New(conf, nil, nil, nil)
			if err == nil || ctrl != nil {
				t.Errorf("got %v, %v, want an error and no controller", ctrl, err)
			}
		})
	}
}
