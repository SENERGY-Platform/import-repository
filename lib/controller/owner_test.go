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

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/import-repository/lib/database"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	permV2Model "github.com/SENERGY-Platform/permissions-v2/pkg/model"
)

// ownerDb keeps documents in memory; methods the move must not use panic through the nil embedded interface.
type ownerDb struct {
	database.Database
	docs      map[string]model.ImportType
	reads     int
	ownerSets int
	failSet   error
}

func (f *ownerDb) GetImportType(_ context.Context, id string) (model.ImportType, bool, error) {
	f.reads++
	it, ok := f.docs[id]
	return it, ok, nil
}

func (f *ownerDb) SetImportTypeOwner(_ context.Context, id string, owner string) (bool, error) {
	f.ownerSets++
	if f.failSet != nil {
		return false, f.failSet
	}
	it, ok := f.docs[id]
	if !ok {
		return false, nil
	}
	it.Owner = owner
	f.docs[id] = it
	return true, nil
}

func (f *ownerDb) ListImportTypes(_ context.Context, _ model.ImportTypeQueryOptions) ([]model.ImportType, int64, error) {
	result := []model.ImportType{}
	for _, it := range f.docs {
		result = append(result, it)
	}
	return result, int64(len(result)), nil
}

// ownerPerm hands out and stores copies, as the real client does through JSON. A stored
// entry with null maps stays null, as permissions-v2 answers it.
type ownerPerm struct {
	permV2.Client
	resources map[string]permV2.ResourcePermissions
	gets      int
	sets      int
	failSet   error
}

func copyPermissions(in permV2.ResourcePermissions) (out permV2.ResourcePermissions) {
	b, err := json.Marshal(in)
	if err == nil {
		err = json.Unmarshal(b, &out)
	}
	if err != nil {
		panic(err)
	}
	return out
}

func (f *ownerPerm) GetResource(token string, topic string, id string) (permV2.Resource, error, int) {
	f.gets++
	if token != permV2.InternalAdminToken || topic != PermV2Topic {
		return permV2.Resource{}, fmt.Errorf("unexpected token or topic %q", topic), http.StatusBadRequest
	}
	r, ok := f.resources[id]
	if !ok {
		return permV2.Resource{}, errors.New("unexpected statuscode 404: not found"), http.StatusNotFound
	}
	return permV2.Resource{Id: id, TopicId: topic, ResourcePermissions: copyPermissions(r)}, nil, http.StatusOK
}

func (f *ownerPerm) SetPermission(token string, topic string, id string, permissions permV2.ResourcePermissions) (permV2.ResourcePermissions, error, int) {
	f.sets++
	if f.failSet != nil {
		return permissions, f.failSet, http.StatusInternalServerError
	}
	if token != permV2.InternalAdminToken || topic != PermV2Topic {
		return permissions, fmt.Errorf("unexpected token or topic %q", topic), http.StatusBadRequest
	}
	f.resources[id] = copyPermissions(permissions)
	return permissions, nil, http.StatusOK
}

type ownerFixture struct {
	ctrl *Controller
	db   *ownerDb
	perm *ownerPerm
}

// it-1 belongs to user-a, is shared with user-b for reading, with a group and with the admin role.
func newOwnerFixture() *ownerFixture {
	f := &ownerFixture{
		db: &ownerDb{docs: map[string]model.ImportType{
			"it-1": {Id: "it-1", Name: "heating", Owner: "user-a", Description: "d", Image: "img", DefaultRestart: true, Cost: 5},
		}},
		perm: &ownerPerm{resources: map[string]permV2.ResourcePermissions{"it-1": storedPermissions()}},
	}
	f.ctrl = &Controller{db: f.db, permV2Client: f.perm}
	return f
}

func storedPermissions() permV2.ResourcePermissions {
	return permV2.ResourcePermissions{
		UserPermissions: map[string]permV2Model.PermissionsMap{
			"user-a": fullRights,
			"user-b": {Read: true},
		},
		GroupPermissions: map[string]permV2Model.PermissionsMap{"/team": {Read: true, Execute: true}},
		RolePermissions:  map[string]permV2Model.PermissionsMap{"admin": fullRights},
	}
}

func (f *ownerFixture) assertUnchanged(t *testing.T) {
	t.Helper()
	if f.perm.sets != 0 || f.db.ownerSets != 0 {
		t.Errorf("wrote permissions %d and owners %d times, want none", f.perm.sets, f.db.ownerSets)
	}
	if !reflect.DeepEqual(f.perm.resources["it-1"], storedPermissions()) {
		t.Errorf("permissions = %+v", f.perm.resources["it-1"])
	}
	if it := f.db.docs["it-1"]; it.Owner != "user-a" {
		t.Errorf("document = %+v", it)
	}
}

func TestSetImportTypeOwnerByANonAdminIsForbiddenBeforeAnyRead(t *testing.T) {
	f := newOwnerFixture()

	_, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", userToken())

	if err == nil || code != http.StatusForbidden {
		t.Fatalf("err = %v, code = %d, want 403", err, code)
	}
	if f.db.reads != 0 || f.perm.gets != 0 {
		t.Errorf("read %d documents and %d permissions before refusing", f.db.reads, f.perm.gets)
	}
	f.assertUnchanged(t)
}

func TestSetImportTypeOwnerRejectsAMissingOwner(t *testing.T) {
	for _, owner := range []string{"", " ", " user-c", "user-c\n"} {
		f := newOwnerFixture()

		_, err, code := f.ctrl.SetImportTypeOwner("it-1", owner, adminToken())

		if err == nil || code != http.StatusBadRequest {
			t.Errorf("owner %q: err = %v, code = %d, want 400", owner, err, code)
		}
		if f.db.reads != 0 || f.perm.gets != 0 {
			t.Errorf("owner %q: read %d documents and %d permissions", owner, f.db.reads, f.perm.gets)
		}
		f.assertUnchanged(t)
	}
}

func TestSetImportTypeOwnerOfAnUnknownIdIsNotFound(t *testing.T) {
	f := newOwnerFixture()

	_, err, code := f.ctrl.SetImportTypeOwner("it-unknown", "user-c", adminToken())

	if err == nil || code != http.StatusNotFound {
		t.Fatalf("err = %v, code = %d, want 404", err, code)
	}
	if f.perm.gets != 0 {
		t.Errorf("asked permissions-v2 %d times about an unknown id", f.perm.gets)
	}
	f.assertUnchanged(t)
}

func TestSetImportTypeOwnerMovesTheImportType(t *testing.T) {
	f := newOwnerFixture()
	before := f.db.docs["it-1"]

	result, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())

	if err != nil || code != http.StatusOK {
		t.Fatalf("err = %v, code = %d", err, code)
	}
	want := storedPermissions()
	want.UserPermissions["user-c"] = fullRights
	if got := f.perm.resources["it-1"]; !reflect.DeepEqual(got, want) {
		t.Errorf("permissions = %+v\nwant %+v", got, want)
	}
	wantDoc := before
	wantDoc.Owner = "user-c"
	if stored := f.db.docs["it-1"]; !reflect.DeepEqual(stored, wantDoc) {
		t.Errorf("document = %+v\nwant only the owner changed: %+v", stored, wantDoc)
	}
	if !reflect.DeepEqual(result, wantDoc) {
		t.Errorf("result = %+v, want %+v", result, wantDoc)
	}
}

func TestSetImportTypeOwnerRaisesPartialRightsOfTheNewOwner(t *testing.T) {
	f := newOwnerFixture()

	_, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-b", adminToken())

	if err != nil || code != http.StatusOK {
		t.Fatalf("err = %v, code = %d", err, code)
	}
	want := storedPermissions()
	want.UserPermissions["user-b"] = fullRights
	if got := f.perm.resources["it-1"]; !reflect.DeepEqual(got, want) {
		t.Errorf("permissions = %+v\nwant %+v", got, want)
	}
}

func TestSetImportTypeOwnerTwiceChangesNothingTheSecondTime(t *testing.T) {
	f := newOwnerFixture()
	first, err, _ := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())
	if err != nil {
		t.Fatal(err)
	}
	permissionsAfterFirst := copyPermissions(f.perm.resources["it-1"])
	f.perm.sets, f.db.ownerSets = 0, 0

	second, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())

	if err != nil || code != http.StatusOK {
		t.Fatalf("err = %v, code = %d", err, code)
	}
	if f.perm.sets != 0 || f.db.ownerSets != 0 {
		t.Errorf("repeat wrote permissions %d and owners %d times", f.perm.sets, f.db.ownerSets)
	}
	if !reflect.DeepEqual(second, first) || !reflect.DeepEqual(f.perm.resources["it-1"], permissionsAfterFirst) {
		t.Errorf("repeat changed the result %+v -> %+v or the permissions", first, second)
	}
}

func TestSetImportTypeOwnerWithoutPermissionsEntryChangesNothing(t *testing.T) {
	f := newOwnerFixture()
	delete(f.perm.resources, "it-1")

	_, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())

	if err == nil || code != http.StatusInternalServerError {
		t.Fatalf("err = %v, code = %d, want 500", err, code)
	}
	if f.perm.sets != 0 || f.db.ownerSets != 0 || f.db.docs["it-1"].Owner != "user-a" {
		t.Errorf("wrote permissions %d and owners %d times", f.perm.sets, f.db.ownerSets)
	}
}

func TestSetImportTypeOwnerKeepsTheDocumentWhenPermissionsFail(t *testing.T) {
	f := newOwnerFixture()
	f.perm.failSet = errors.New("permissions-v2 down")

	_, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())

	if err == nil || code != http.StatusInternalServerError {
		t.Fatalf("err = %v, code = %d, want 500", err, code)
	}
	if f.db.ownerSets != 0 || f.db.docs["it-1"].Owner != "user-a" {
		t.Errorf("document owner set %d times to %q after the permissions failed", f.db.ownerSets, f.db.docs["it-1"].Owner)
	}
}

func TestSetImportTypeOwnerRepeatCompletesAfterTheDocumentWriteFailed(t *testing.T) {
	f := newOwnerFixture()
	f.db.failSet = errors.New("mongo down")

	_, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())

	if err == nil || code != http.StatusInternalServerError {
		t.Fatalf("err = %v, code = %d, want 500", err, code)
	}
	if f.perm.resources["it-1"].UserPermissions["user-c"] != fullRights || f.db.docs["it-1"].Owner != "user-a" {
		t.Fatalf("after the failure: permissions %+v, owner %q", f.perm.resources["it-1"], f.db.docs["it-1"].Owner)
	}

	f.db.failSet = nil
	f.perm.sets = 0
	result, err, code := f.ctrl.SetImportTypeOwner("it-1", "user-c", adminToken())

	if err != nil || code != http.StatusOK || result.Owner != "user-c" || f.db.docs["it-1"].Owner != "user-c" {
		t.Fatalf("repeat: err = %v, code = %d, owner %q", err, code, f.db.docs["it-1"].Owner)
	}
	if f.perm.sets != 0 {
		t.Errorf("repeat wrote the permissions again %d times", f.perm.sets)
	}
}

// permissions-v2 can answer an entry whose role, user and group maps are null.
func TestMigrateSurvivesAnEntryWithNullPermissionMaps(t *testing.T) {
	f := newOwnerFixture()
	f.perm.resources["it-1"] = permV2.ResourcePermissions{}

	err := f.ctrl.Migrate()

	if err != nil {
		t.Fatal(err)
	}
	got := f.perm.resources["it-1"]
	if got.UserPermissions["user-a"] != fullRights || got.RolePermissions["admin"] != fullRights {
		t.Errorf("permissions = %+v", got)
	}
}
