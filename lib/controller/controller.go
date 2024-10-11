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
	"context"
	"time"

	deviceRepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/database"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func New(config config.Config, db database.Database, permV2Client permV2.Client) (ctrl *Controller, err error) {
	ctrl = &Controller{
		db:               db,
		config:           config,
		permV2Client:     permV2Client,
		deviceRepoClient: deviceRepo.NewClient(config.DeviceRepoUrl),
	}
	_, err, _ = ctrl.permV2Client.SetTopic(permV2.InternalAdminToken, permV2.Topic{
		Id: PermV2Topic,
		DefaultPermissions: permV2.ResourcePermissions{
			RolePermissions: map[string]permV2.PermissionsMap{
				"admin": {
					Read:         true,
					Write:        true,
					Execute:      true,
					Administrate: true,
				},
			},
		},
	})
	return
}

type Controller struct {
	db               database.Database
	config           config.Config
	permV2Client     permV2.Client
	deviceRepoClient deviceRepo.Interface
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
