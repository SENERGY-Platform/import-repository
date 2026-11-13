/*
 * Copyright 2020 InfAI (CC SES)
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
	"errors"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"net/http"
	"strings"

	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	permV2Model "github.com/SENERGY-Platform/permissions-v2/pkg/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/hashicorp/go-uuid"
)

const idPrefix = "urn:infai:ses:import-type:"
const PermV2Topic = "import-types"

func (this *Controller) CreateImportType(importType model.ImportType, token jwt.Token) (result model.ImportType, err error, code int) {
	id, err := uuid.GenerateUUID()
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if importType.Id != "" {
		return result, errors.New("explicit setting of id not allowed"), http.StatusBadRequest
	}
	importType.Id = idPrefix + id
	if importType.Owner != "" {
		return result, errors.New("explicit setting of owner not allowed"), http.StatusBadRequest
	}
	importType.Owner = token.GetUserId()
	model.SetContentVariableAspectIdsOnWrite(&importType)
	if this.config.Validate {
		err, code = this.ValidateImportType(token, importType)
		if err != nil {
			return result, err, code
		}
	}
	ctx, _ := getTimeoutContext()
	err = this.db.SetImportType(ctx, importType)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	_, err, code = this.permV2Client.SetPermission(client.InternalAdminToken, PermV2Topic, importType.Id, client.ResourcePermissions{
		UserPermissions: map[string]permV2Model.PermissionsMap{
			importType.Owner: {
				Read:         true,
				Write:        true,
				Execute:      true,
				Administrate: true,
			},
		},
		RolePermissions: map[string]permV2Model.PermissionsMap{
			"admin": {
				Read:         true,
				Write:        true,
				Execute:      true,
				Administrate: true,
			},
		},
	})
	if err != nil {
		// Without a permissions entry the stored type is invisible in lists, and a later migration
		// would give it rights again next to the client's retry.
		rollbackCtx, _ := getTimeoutContext()
		if rollbackErr := this.db.RemoveImportType(rollbackCtx, importType.Id); rollbackErr != nil {
			log.Logger.Error("unable to remove an import type whose permissions could not be set", "import_type_id", importType.Id, attributes.ErrorKey, rollbackErr)
			err = errors.Join(err, rollbackErr)
		}
		return result, err, code
	}
	model.SetContentVariableAspectIdsOnRead(&importType)
	return importType, nil, http.StatusCreated
}

func (this *Controller) ReadImportType(id string, token jwt.Token) (result model.ImportType, err error, errCode int) {
	err, code := this.CheckAccessToImportType(token, id, permV2Model.Read)
	if err != nil {
		result = model.ImportType{}
		return result, err, code
	}
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetImportType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	model.SetContentVariableAspectIdsOnRead(&result)
	return result, nil, http.StatusOK
}

func (this *Controller) ListImportTypes(token jwt.Token, options model.ImportTypeListOptions) (result []model.ImportType, total int64, err error, errCode int) {
	ids := []string{}
	if options.Ids == nil {
		if token.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permV2Client.ListAccessibleResourceIds(token.Token, PermV2Topic, permV2Model.ListOptions{}, permV2Model.Read)
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			if ids == nil {
				//an empty id filter, not the absent id filter of an admin
				ids = []string{}
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permV2Client.CheckMultiplePermissions(token.Token, PermV2Topic, options.Ids, permV2Model.Read)
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}
	options.Ids = ids
	queryOptions, err := this.importTypeQueryOptions(options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListImportTypes(ctx, queryOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	model.SetContentVariableAspectIdsOnReadList(result)
	return result, total, nil, http.StatusOK
}

// importTypeQueryOptions resolves the aspects of every filter-criteria for the database. By
// default the AspectIds of a criteria are alternatives and the caller names the aspect subtree
// it wants itself. With AndCombineCriteriaAspectIds every named aspect has to be carried by
// the same content variable, and each of them covers its subtree, which is resolved here
// because only the device-repository knows the aspect hierarchy.
func (this *Controller) importTypeQueryOptions(options model.ImportTypeListOptions) (result model.ImportTypeQueryOptions, err error) {
	result = model.ImportTypeQueryOptions{
		Ids:    options.Ids,
		Search: options.Search,
		Limit:  options.Limit,
		Offset: options.Offset,
		SortBy: options.SortBy,
	}
	if options.Criteria == nil {
		return result, nil
	}
	result.Criteria = []model.ImportTypeCriteriaQuery{}
	for _, criteria := range options.Criteria {
		query := model.ImportTypeCriteriaQuery{FunctionId: criteria.FunctionId}
		aspectIds := model.FilterCriteriaAspectIds(criteria)
		if options.AndCombineCriteriaAspectIds {
			for _, aspectId := range aspectIds {
				if aspectId == "" {
					continue
				}
				subtree, err := this.aspectIdWithDescendents(aspectId)
				if err != nil {
					return result, err
				}
				query.AspectIdSets = append(query.AspectIdSets, subtree)
			}
		} else if len(aspectIds) > 0 {
			query.AspectIdSets = [][]string{aspectIds}
		}
		result.Criteria = append(result.Criteria, query)
	}
	return result, nil
}

func (this *Controller) SetImportType(importType model.ImportType, token jwt.Token) (err error, errCode int) {
	err, code := this.CheckAccessToImportType(token, importType.Id, permV2Model.Write)
	if err != nil {
		return err, code
	}
	ctx, _ := getTimeoutContext()
	existing, exists, err := this.db.GetImportType(ctx, importType.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return errors.New("not found"), http.StatusNotFound
	}
	if importType.Owner != existing.Owner {
		return errors.New("transfer of ownership not possible!"), http.StatusBadRequest
	}
	model.SetContentVariableAspectIdsOnWrite(&importType)
	if this.config.Validate {
		err, code = this.ValidateImportType(token, importType)
		if err != nil {
			return err, code
		}
	}
	ctx, _ = getTimeoutContext()
	err = this.db.SetImportType(ctx, importType)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

var fullRights = permV2Model.PermissionsMap{Read: true, Write: true, Execute: true, Administrate: true}

// SetImportTypeOwner moves the import type to owner, for administrators only: the owner gets full
// rights in permissions-v2 first, every other entry stays as stored, then the document's owner is set.
// If the document write fails the owner merely holds extra rights and a repeat completes the move.
// Permissions are written back without a version check, so a permission edit made in between is lost.
func (this *Controller) SetImportTypeOwner(id string, owner string, token jwt.Token) (result model.ImportType, err error, code int) {
	if !token.IsAdmin() {
		return result, errors.New("admin role required"), http.StatusForbidden
	}
	if owner == "" || strings.TrimSpace(owner) != owner {
		return result, errors.New("owner must be a user id without surrounding whitespace"), http.StatusBadRequest
	}
	ctx, cancel := getTimeoutContext()
	defer cancel()
	existing, exists, err := this.db.GetImportType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}

	resource, err, _ := this.permV2Client.GetResource(client.InternalAdminToken, PermV2Topic, id)
	if err != nil {
		// Also when permissions-v2 knows no entry: an entry written here could keep the creator of an import type that is just being created out.
		return result, err, http.StatusInternalServerError
	}
	permissions := resource.ResourcePermissions
	if permissions.UserPermissions[owner] != fullRights {
		if permissions.UserPermissions == nil {
			permissions.UserPermissions = map[string]permV2Model.PermissionsMap{}
		}
		permissions.UserPermissions[owner] = fullRights
		_, err, _ = this.permV2Client.SetPermission(client.InternalAdminToken, PermV2Topic, id, permissions)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
	}

	if existing.Owner != owner {
		found, err := this.db.SetImportTypeOwner(ctx, id, owner)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		if !found {
			return result, errors.New("not found"), http.StatusNotFound
		}
	}
	result, exists, err = this.db.GetImportType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	model.SetContentVariableAspectIdsOnRead(&result)
	return result, nil, http.StatusOK
}

func (this *Controller) DeleteImportType(id string, token jwt.Token) (err error, errCode int) {
	return this.DeleteImportTypeWithOptions(id, token, model.DeleteOptions{})
}

// DeleteImportTypeWithOptions refuses an import type that smart service releases or import
// instances still use, with an *model.InUseError and 409, unless opts.Force is set by an
// administrator. When the usage cannot be determined nothing is deleted (502), also with force.
// Check and delete are not atomic: an instance created in between is not seen.
func (this *Controller) DeleteImportTypeWithOptions(id string, token jwt.Token, opts model.DeleteOptions) (err error, errCode int) {
	if opts.Force && !token.IsAdmin() {
		return errors.New("force needs the admin role"), http.StatusForbidden
	}
	err, code := this.CheckAccessToImportType(token, id, permV2Model.Administrate)
	if err != nil {
		return err, code
	}
	used, err := this.usageOf(id, token)
	if err != nil {
		return err, http.StatusBadGateway
	}
	if used.Used() && !opts.Force {
		return &model.InUseError{Usage: used}, http.StatusConflict
	}
	err, code = this.deleteImportType(id)
	if err != nil {
		return err, code
	}
	if used.Used() {
		log.Logger.Warn("deleted an import type that is still used", "import_type_id", id, "user_id", token.GetUserId(),
			"releases", used.SmartServices.Releases, "release_instances", used.SmartServices.Instances, "import_instances", used.Imports.Instances)
	}
	return nil, code
}

// usageOf asks both sources with the caller's token. Either failing is model.ErrUsageUnavailable:
// the detail may hold internals, so it is logged and not returned.
func (this *Controller) usageOf(id string, token jwt.Token) (used model.ImportTypeUsage, err error) {
	ctx, cancel := getTimeoutContext()
	defer cancel()
	used.SmartServices, err = this.usage.SmartServices(ctx, id, token.Token)
	if err != nil {
		log.Logger.Warn("could not ask the smart service repository for the usage of an import type", "import_type_id", id, attributes.ErrorKey, err)
		return used, model.ErrUsageUnavailable
	}
	used.Imports, err = this.usage.Imports(ctx, id, token.Token)
	if err != nil {
		log.Logger.Warn("could not ask import-deploy for the usage of an import type", "import_type_id", id, attributes.ErrorKey, err)
		return used, model.ErrUsageUnavailable
	}
	return used, nil
}

func (this *Controller) deleteImportType(id string) (err error, code int) {
	err, code = this.permV2Client.RemoveResource(client.InternalAdminToken, PermV2Topic, id)
	if err != nil {
		return err, code
	}
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveImportType(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusNoContent
}

func (this *Controller) CheckAccessToImportType(token jwt.Token, id string, action permV2Model.Permission) (err error, errCode int) {
	ok, err, errCode := this.permV2Client.CheckPermission(token.Token, PermV2Topic, id, action)
	if err != nil {
		return err, errCode
	}
	if !ok {
		return errors.New("forbidden"), http.StatusForbidden
	}
	return
}
