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
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/auth"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"net/http"
)

func (c Client) ReadImportType(id string, jwt auth.Token) (result model.ImportType, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/import-types/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+jwt.Jwt())
	return do[model.ImportType](req)
}

// CheckAccessToImportType NOT IMPLEMENTED!
func (c Client) CheckAccessToImportType(jwt auth.Token, id string, action model.AuthAction) (err error, code int) {
	return errors.New("not implemented"), http.StatusInternalServerError
}

func (c Client) CreateImportType(importType model.ImportType, jwt auth.Token) (result model.ImportType, err error, code int) {
	b, err := json.Marshal(importType)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/import-types", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+jwt.Jwt())
	return do[model.ImportType](req)
}

func (c Client) SetImportType(importType model.ImportType, jwt auth.Token) (err error, code int) {
	b, err := json.Marshal(importType)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/import-types", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+jwt.Jwt())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}

func (c Client) DeleteImportType(id string, jwt auth.Token) (err error, errCode int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/import-types/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+jwt.Jwt())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}
