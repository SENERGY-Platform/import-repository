/*
 * Copyright 2026 InfAI (CC SES)
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
	"errors"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func init() {
	endpoints = append(endpoints, UsersEndpoints)
}

type usersHandler struct {
	control Controller
}

func UsersEndpoints(config config.Config, control Controller, router *gin.Engine) {
	handler := usersHandler{control: control}
	router.DELETE("/users/:id", handler.deleteUser)
}

// deleteUser godoc
// @Summary Delete user
// @Description Administrators only; user-management calls this while deleting a user. Deletes the import types the user administers alone and removes the user's permissions from all others. Does not check whether smart services or import instances still use the import types.
// @Tags users
// @Param id path string true "User id"
// @Success 204
// @Failure 400 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Security Bearer
// @Router /users/{id} [delete]
func (handler usersHandler) deleteUser(c *gin.Context) {
	token, err := jwt.GetParsedToken(c.Request)
	if err != nil {
		_ = c.Error(errors.Join(model.ErrBadRequest, err))
		return
	}
	err, code := handler.control.DeleteUser(c.Param("id"), token)
	if err != nil {
		_ = c.Error(errors.Join(model.GetError(code), err))
		return
	}
	c.Status(code)
}
