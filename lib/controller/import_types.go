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
	"github.com/SENERGY-Platform/import-repository/lib/auth"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/hashicorp/go-uuid"
	"net/http"
)

const idPrefix = "urn:infai:ses:import-type:"

func (this *Controller) CreateImportType(importType model.ImportType, jwt auth.Token) (result model.ImportType, err error, code int) {
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
	importType.Owner = jwt.GetUserId()
	if this.config.Validate {
		err, code = this.ValidateImportType(jwt, importType)
		if err != nil {
			return result, err, code
		}
	}
	err = this.producer.PublishImportType(importType, importType.Owner)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	return importType, nil, http.StatusCreated
}

func (this *Controller) ReadImportType(id string, jwt auth.Token) (result model.ImportType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetImportType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	err, code := this.CheckAccessToImportType(jwt, id, model.READ)
	if err != nil {
		result = model.ImportType{}
		return result, err, code
	}

	return result, nil, http.StatusOK
}

func (this *Controller) SetImportType(importType model.ImportType, jwt auth.Token) (err error, errCode int) {
	ctx, _ := getTimeoutContext()
	existing, exists, err := this.db.GetImportType(ctx, importType.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return errors.New("not found"), http.StatusNotFound
	}
	err, code := this.CheckAccessToImportType(jwt, importType.Id, model.WRITE)
	if err != nil {
		return err, code
	}
	if importType.Owner != existing.Owner {
		return errors.New("transfer of ownership not possible!"), http.StatusBadRequest
	}
	if this.config.Validate {
		err, code = this.ValidateImportType(jwt, importType)
		if err != nil {
			return err, code
		}
	}
	err = this.producer.PublishImportType(importType, importType.Owner)
	if err != nil {
		return err, http.StatusInternalServerError // TODO rollback
	}

	return nil, http.StatusOK
}

func (this *Controller) DeleteImportType(id string, jwt auth.Token) (err error, errCode int) {
	ctx, _ := getTimeoutContext()
	existing, exists, err := this.db.GetImportType(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return errors.New("not found"), http.StatusNotFound
	}
	err, code := this.CheckAccessToImportType(jwt, id, model.WRITE)
	if err != nil {
		return err, code
	}
	err = this.producer.PublishImportTypeDelete(id, existing.Owner)
	if err != nil {
		return err, http.StatusInternalServerError // TODO rollback
	}
	return nil, http.StatusNoContent
}

func (this *Controller) DeleteImportTypeFromDB(id string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveImportType(ctx, id)
}

func (this *Controller) SetImportTypeInDB(importType model.ImportType) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetImportType(ctx, importType)
}

func (this *Controller) CheckAccessToImportType(jwt auth.Token, id string, action model.AuthAction) (err error, errCode int) {
	ok, err := this.security.CheckBool(jwt, this.config.ImportTypeTopic, id, action)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	return
}
