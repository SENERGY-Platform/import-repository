/*
 * Copyright 2023 InfAI (CC SES)
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

package client

import (
	"bytes"
	"encoding/json"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"strconv"

	"github.com/SENERGY-Platform/import-repository/lib/model"
)

func (c Client) ReadImportType(id string, token jwt.Token) (result model.ImportType, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/import-types/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", "Bearer "+token.Jwt())
	return do[model.ImportType](req)
}

func (c Client) ListImportTypes(token jwt.Token, limit int64, offset int64, sort string) (result []model.ImportType, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/import-types?limit="+strconv.FormatInt(limit, 10)+"&offset="+strconv.FormatInt(offset, 10)+"&sort="+sort, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", "Bearer "+token.Jwt())
	return do[[]model.ImportType](req)
}

func (c Client) ListImportTypesV2(token jwt.Token, options model.ImportTypeListOptions) (result []model.ImportType, err error, errCode int) {
	//TODO implement me
	panic("implement me")
}

func (c Client) CreateImportType(importType model.ImportType, token jwt.Token) (result model.ImportType, err error, code int) {
	b, err := json.Marshal(importType)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/import-types", bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", "Bearer "+token.Jwt())
	return do[model.ImportType](req)
}

func (c Client) SetImportType(importType model.ImportType, token jwt.Token) (err error, code int) {
	b, err := json.Marshal(importType)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/import-types", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", "Bearer "+token.Jwt())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}

func (c Client) DeleteImportType(id string, token jwt.Token) (err error, errCode int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/import-types/"+id, nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", "Bearer "+token.Jwt())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}
