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
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func init() {
	endpoints = append(endpoints, ImportTypesEndpoints)
}

func ImportTypesEndpoints(config config.Config, control Controller, router *gin.Engine) {
	resource := "/import-types"

	router.GET(resource, func(c *gin.Context) {
		token, err := jwt.GetParsedToken(c.Request)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}

		listOptions := model.ImportTypeListOptions{
			Limit:  100,
			Offset: 0,
		}
		limitParam := c.Query("limit")
		if limitParam != "" {
			listOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, errors.New("unable to parse limit"), err))
			return
		}

		offsetParam := c.Query("offset")
		if offsetParam != "" {
			listOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, errors.New("unable to parse offset"), err))
			return
		}

		idsParam := c.Query("ids")
		if _, hasIds := c.GetQuery("ids"); hasIds {
			if idsParam != "" {
				listOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				listOptions.Ids = []string{}
			}
		}

		criteria := c.Query("criteria")
		if criteria != "" {
			listOptions.Criteria = []model.ImportTypeFilterCriteria{}
			err = json.Unmarshal([]byte(criteria), &listOptions.Criteria)
			if err != nil {
				_ = c.Error(errors.Join(model.ErrBadRequest, err))
				return
			}
		}

		listOptions.Search = c.Query("search")
		listOptions.SortBy = c.Query("sort")
		if listOptions.SortBy == "" {
			listOptions.SortBy = "name.asc"
		}

		result, total, err, errCode := control.ListImportTypes(token, listOptions)
		if err != nil {
			_ = c.Error(errors.Join(model.GetError(errCode), err))
			return
		}
		c.Header("X-Total-Count", strconv.FormatInt(total, 10))
		c.JSON(http.StatusOK, result)
	})

	router.GET(resource+"/:id", func(c *gin.Context) {
		id := c.Param("id")
		token, err := jwt.GetParsedToken(c.Request)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}
		result, err, errCode := control.ReadImportType(id, token)
		if err != nil {
			_ = c.Error(errors.Join(model.GetError(errCode), err))
			return
		}
		c.JSON(http.StatusOK, result)
	})

	router.DELETE(resource+"/:id", func(c *gin.Context) {
		id := c.Param("id")
		token, err := jwt.GetParsedToken(c.Request)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}
		err, errCode := control.DeleteImportType(id, token)
		if err != nil {
			_ = c.Error(errors.Join(model.GetError(errCode), err))
			return
		}
		c.Status(errCode)
	})

	router.PUT(resource+"/:id", func(c *gin.Context) {
		id := c.Param("id")
		token, err := jwt.GetParsedToken(c.Request)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}
		importType := model.ImportType{}
		err = c.ShouldBind(&importType)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}
		if id != importType.Id {
			_ = c.Error(errors.Join(model.ErrBadRequest, errors.New("IDs don't match")))
			return
		}
		err, code := control.SetImportType(importType, token)
		if err != nil {
			_ = c.Error(errors.Join(model.GetError(code), err))
			return
		}
		c.Status(http.StatusOK)
	})

	router.POST(resource, func(c *gin.Context) {
		importType := model.ImportType{}
		token, err := jwt.GetParsedToken(c.Request)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}
		err = c.ShouldBind(&importType)
		if err != nil {
			_ = c.Error(errors.Join(model.ErrBadRequest, err))
			return
		}
		result, err, code := control.CreateImportType(importType, token)
		if err != nil {
			_ = c.Error(errors.Join(model.GetError(code), err))
			return
		}
		c.JSON(code, result)
	})
}
