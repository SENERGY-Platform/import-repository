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
	"net/url"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/import-repository/lib/model"
)

func (c Client) ReadImportType(id string, token jwt.Token) (result model.ImportType, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/import-types/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	return do[model.ImportType](req)
}

func (c Client) ListImportTypes(token jwt.Token, options model.ImportTypeListOptions) (result []model.ImportType, total int64, err error, errCode int) {
	queryString := ""
	query := url.Values{}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.Ids != nil {
		query.Set("ids", strings.Join(options.Ids, ","))
	}
	if options.SortBy != "" {
		query.Set("sort", options.SortBy)
	}
	if options.Limit != 0 {
		query.Set("limit", strconv.FormatInt(options.Limit, 10))
	}
	if options.Offset != 0 {
		query.Set("offset", strconv.FormatInt(options.Offset, 10))
	}
	if len(options.Criteria) > 0 {
		filterStr, err := json.Marshal(options.Criteria)
		if err != nil {
			return result, total, err, http.StatusBadRequest
		}
		query.Add("criteria", string(filterStr))
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/import-types"+queryString, nil)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	return doWithTotalInResult[[]model.ImportType](req)
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
	req.Header.Set("Authorization", token.Jwt())
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
	req.Header.Set("Authorization", token.Jwt())
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
	req.Header.Set("Authorization", token.Jwt())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}
