/*
 * Copyright 2022 InfAI (CC SES)
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

package com

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/auth"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"net/http"
	"runtime/debug"
	"strconv"
)

type AspectNodeQuery struct {
	Ids []string `json:"ids"`
}

func (this *Com) GetAspectNodes(ids []string, token auth.Token) ([]model.AspectNode, error) {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(AspectNodeQuery{Ids: ids})
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	req, err := http.NewRequest("POST", this.config.DeviceRepoUrl+"/query/aspect-nodes", requestBody)
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	req.Header.Set("Authorization", token.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, errors.New("unexpected status code " + strconv.Itoa(resp.StatusCode))
	}

	nodes := []model.AspectNode{}
	err = json.NewDecoder(resp.Body).Decode(&nodes)
	return nodes, err
}
