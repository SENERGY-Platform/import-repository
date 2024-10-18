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
	"net/http"

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
	return result, nil, http.StatusOK
}

func (this *Controller) ListImportTypes(token jwt.Token, limit int64, offset int64, sort string) (result []model.ImportType, err error, errCode int) {
	ids, err, errCode := this.permV2Client.ListAccessibleResourceIds(token.Token, PermV2Topic, permV2Model.ListOptions{}, permV2Model.Read)
	if err != nil {
		return
	}
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListImportTypes(ctx, limit, offset, sort, ids)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ListImportTypesV2(token jwt.Token, options model.ImportTypeListOptions) (result []model.ImportType, err error, errCode int) {
	//TODO: implement
	panic("implement me")
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

func (this *Controller) DeleteImportType(id string, token jwt.Token) (err error, errCode int) {
	err, code := this.CheckAccessToImportType(token, id, permV2Model.Administrate)
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
