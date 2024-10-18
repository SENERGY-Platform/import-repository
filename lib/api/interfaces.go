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
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

type Controller interface {
	ReadImportType(id string, token jwt.Token) (result model.ImportType, err error, errCode int)
	ListImportTypes(token jwt.Token, limit int64, offset int64, sort string) (result []model.ImportType, err error, errCode int)
	ListImportTypesV2(token jwt.Token, options model.ImportTypeListOptions) (result []model.ImportType, err error, errCode int)
	CreateImportType(importType model.ImportType, token jwt.Token) (result model.ImportType, err error, code int)
	SetImportType(importType model.ImportType, token jwt.Token) (err error, code int)
	DeleteImportType(id string, token jwt.Token) (err error, errCode int)
}
