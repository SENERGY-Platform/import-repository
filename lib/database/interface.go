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

package database

import (
	"context"
	"github.com/SENERGY-Platform/import-repository/lib/model"
)

type Database interface {
	GetImportType(ctx context.Context, id string) (device model.ImportType, exists bool, err error)
	ListImportTypes(ctx context.Context, limit int64, offset int64, sort string, limitToIds []string) (result []model.ImportType, err error)
	SetImportType(ctx context.Context, importType model.ImportType) error
	RemoveImportType(ctx context.Context, id string) error
}
