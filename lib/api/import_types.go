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

type importTypesHandler struct {
	control Controller
}

func ImportTypesEndpoints(config config.Config, control Controller, router *gin.Engine) {
	resource := "/import-types"
	handler := importTypesHandler{control: control}

	router.GET(resource, handler.listImportTypes)
	router.GET(resource+"/:id", handler.readImportType)
	router.DELETE(resource+"/:id", handler.deleteImportType)
	router.PUT(resource+"/:id", handler.setImportType)
	router.POST(resource, handler.createImportType)
}

// listImportTypes godoc
// @Summary List import types
// @Description Returns import types visible to the caller. If `ids` is provided, pagination is ignored.
// @Tags import-types
// @Produce json
// @Param limit query int false "Maximum number of results" default(100)
// @Param offset query int false "Result offset" default(0)
// @Param ids query string false "Comma-separated import type ids"
// @Param criteria query string false "JSON-encoded filter criteria array"
// @Param search query string false "Free-text search term"
// @Param sort query string false "Sort order" default(name.asc)
// @Success 200 {array} model.ImportType
// @Header 200 {integer} X-Total-Count "Total number of matching import types"
// @Failure 400 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Security Bearer
// @Router /import-types [get]
func (handler importTypesHandler) listImportTypes(c *gin.Context) {
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

	result, total, err, errCode := handler.control.ListImportTypes(token, listOptions)
	if err != nil {
		_ = c.Error(errors.Join(model.GetError(errCode), err))
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, result)
}

// readImportType godoc
// @Summary Get import type
// @Description Returns a single import type by id.
// @Tags import-types
// @Produce json
// @Param id path string true "Import type id"
// @Success 200 {object} model.ImportType
// @Failure 400 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 404 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Security Bearer
// @Router /import-types/{id} [get]
func (handler importTypesHandler) readImportType(c *gin.Context) {
	id := c.Param("id")
	token, err := jwt.GetParsedToken(c.Request)
	if err != nil {
		_ = c.Error(errors.Join(model.ErrBadRequest, err))
		return
	}
	result, err, errCode := handler.control.ReadImportType(id, token)
	if err != nil {
		_ = c.Error(errors.Join(model.GetError(errCode), err))
		return
	}
	c.JSON(http.StatusOK, result)
}

// deleteImportType godoc
// @Summary Delete import type
// @Description Deletes an import type by id.
// @Tags import-types
// @Param id path string true "Import type id"
// @Success 200
// @Failure 400 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 404 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Security Bearer
// @Router /import-types/{id} [delete]
func (handler importTypesHandler) deleteImportType(c *gin.Context) {
	id := c.Param("id")
	token, err := jwt.GetParsedToken(c.Request)
	if err != nil {
		_ = c.Error(errors.Join(model.ErrBadRequest, err))
		return
	}
	err, errCode := handler.control.DeleteImportType(id, token)
	if err != nil {
		_ = c.Error(errors.Join(model.GetError(errCode), err))
		return
	}
	c.Status(errCode)
}

// setImportType godoc
// @Summary Update import type
// @Description Replaces an import type. The request body id must match the path id.
// @Tags import-types
// @Accept json
// @Param id path string true "Import type id"
// @Param importType body model.ImportType true "Full import type payload"
// @Success 200
// @Failure 400 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 404 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Security Bearer
// @Router /import-types/{id} [put]
func (handler importTypesHandler) setImportType(c *gin.Context) {
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
	err, code := handler.control.SetImportType(importType, token)
	if err != nil {
		_ = c.Error(errors.Join(model.GetError(code), err))
		return
	}
	c.Status(http.StatusOK)
}

// createImportType godoc
// @Summary Create import type
// @Description Creates a new import type.
// @Tags import-types
// @Accept json
// @Produce json
// @Param importType body model.ImportType true "Import type payload"
// @Success 200 {object} model.ImportType
// @Failure 400 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Security Bearer
// @Router /import-types [post]
func (handler importTypesHandler) createImportType(c *gin.Context) {
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
	result, err, code := handler.control.CreateImportType(importType, token)
	if err != nil {
		_ = c.Error(errors.Join(model.GetError(code), err))
		return
	}
	c.JSON(code, result)
}
