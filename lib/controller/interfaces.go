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
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SmartEnergyPlatform/jwt-http-router"
)

type Security interface {
	CheckBool(jwt jwt_http_router.Jwt, kind string, id string, action model.AuthAction) (allowed bool, err error)
	CheckMultiple(jwt jwt_http_router.Jwt, kind string, ids []string, action model.AuthAction) (map[string]bool, error)
	GetAsUser(jwt jwt_http_router.Jwt, url string, result *[]interface{}) (err error)
}

type Producer interface {
	PublishImportTypeDelete(id string, owner string) error
	PublishImportType(hub model.ImportType, owner string) (err error)
	PublishDeleteUserRights(resource string, id string, userId string) error
}
