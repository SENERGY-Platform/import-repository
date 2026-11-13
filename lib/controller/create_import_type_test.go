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
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/database"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

// createFake stores documents in memory and fails the permissions write on demand; the embedded
// interfaces stay nil because a create must touch nothing else.
type createFake struct {
	database.Database
	permV2.Client
	stored    map[string]model.ImportType
	permErr   error
	removeErr error
}

func (f *createFake) SetImportType(_ context.Context, importType model.ImportType) error {
	f.stored[importType.Id] = importType
	return nil
}

func (f *createFake) RemoveImportType(_ context.Context, id string) error {
	if f.removeErr != nil {
		return f.removeErr
	}
	delete(f.stored, id)
	return nil
}

func (f *createFake) SetPermission(_ string, _ string, _ string, _ permV2.ResourcePermissions) (permV2.ResourcePermissions, error, int) {
	if f.permErr != nil {
		return permV2.ResourcePermissions{}, f.permErr, http.StatusBadGateway
	}
	return permV2.ResourcePermissions{}, nil, http.StatusOK
}

func newCreateController(f *createFake) *Controller {
	f.stored = map[string]model.ImportType{}
	return &Controller{db: f, permV2Client: f}
}

func TestCreateImportTypeRemovesTheStoredTypeWhenPermissionsFail(t *testing.T) {
	permErr := errors.New("permissions-v2 unreachable")
	f := &createFake{permErr: permErr}
	ctrl := newCreateController(f)
	_, err, code := ctrl.CreateImportType(model.ImportType{Name: "first"}, userToken())
	if !errors.Is(err, permErr) || code != http.StatusBadGateway {
		t.Fatalf("got %v, %d, want the permissions error and 502", err, code)
	}
	if len(f.stored) != 0 {
		t.Errorf("import type left behind without permissions: %v", f.stored)
	}
}

func TestCreateImportTypeReturnsBothErrorsWhenTheRollbackFails(t *testing.T) {
	permErr := errors.New("permissions-v2 unreachable")
	removeErr := errors.New("mongo unreachable")
	f := &createFake{permErr: permErr, removeErr: removeErr}
	ctrl := newCreateController(f)
	_, err, code := ctrl.CreateImportType(model.ImportType{Name: "first"}, userToken())
	if !errors.Is(err, permErr) || !errors.Is(err, removeErr) || code != http.StatusBadGateway {
		t.Fatalf("got %v, %d, want both errors and 502", err, code)
	}
}

func TestCreateImportTypeKeepsTheStoredTypeOnSuccess(t *testing.T) {
	f := &createFake{}
	ctrl := newCreateController(f)
	result, err, code := ctrl.CreateImportType(model.ImportType{Name: "first"}, userToken())
	if err != nil || code != http.StatusCreated {
		t.Fatalf("got %v, %d", err, code)
	}
	if _, ok := f.stored[result.Id]; !ok {
		t.Errorf("created import type %s is not stored", result.Id)
	}
}
