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

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/julienschmidt/httprouter"
)

func init() {
	endpoints = append(endpoints, ImportTypesEndpoints)
}

func ImportTypesEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/import-types"

	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		listOptions := model.ImportTypeListOptions{
			Limit:  100,
			Offset: 0,
		}
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			listOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			listOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				listOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				listOptions.Ids = []string{}
			}
		}

		criteria := request.URL.Query().Get("criteria")
		if criteria != "" {
			listOptions.Criteria = []model.ImportTypeFilterCriteria{}
			err = json.Unmarshal([]byte(criteria), &listOptions.Criteria)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		listOptions.Search = request.URL.Query().Get("search")
		listOptions.SortBy = request.URL.Query().Get("sort")
		if listOptions.SortBy == "" {
			listOptions.SortBy = "name.asc"
		}

		result, total, err, errCode := control.ListImportTypes(token, listOptions)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.ReadImportType(id, token)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.DELETE(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, errCode := control.DeleteImportType(id, token)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(errCode)
		return
	})

	router.PUT(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		importType := model.ImportType{}
		err = json.NewDecoder(request.Body).Decode(&importType)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if id != importType.Id {
			http.Error(writer, "IDs don't match", http.StatusBadRequest)
			return
		}
		err, code := control.SetImportType(importType, token)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.POST(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		importType := model.ImportType{}
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.NewDecoder(request.Body).Decode(&importType)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, code := control.CreateImportType(importType, token)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(code)
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
			return
		}
		return
	})
}
