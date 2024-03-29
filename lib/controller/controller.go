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
	"github.com/SENERGY-Platform/import-repository/lib/com"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/database"
	"time"
)

func New(config config.Config, db database.Database, security Security, producer Producer) (ctrl *Controller, err error) {
	ctrl = &Controller{
		db:       db,
		producer: producer,
		security: security,
		config:   config,
		com:      com.New(config),
	}
	return
}

type Controller struct {
	db       database.Database
	security Security
	producer Producer
	com      *com.Com
	config   config.Config
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
