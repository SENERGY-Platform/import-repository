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
	"github.com/SENERGY-Platform/import-repository/lib/auth"
	"github.com/SENERGY-Platform/import-repository/lib/model"
)

type Controller interface {
	ReadImportType(id string, jwt auth.Token) (result model.ImportType, err error, errCode int)
	CheckAccessToImportType(jwt auth.Token, id string, action model.AuthAction) (err error, code int)
	CreateImportType(importType model.ImportType, jwt auth.Token) (result model.ImportType, err error, code int)
	SetImportType(importType model.ImportType, jwt auth.Token) (err error, code int)
	DeleteImportType(id string, jwt auth.Token) (err error, errCode int)
}
