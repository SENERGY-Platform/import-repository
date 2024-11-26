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

package lib

import (
	"context"
	"log"
	"sync"

	"github.com/SENERGY-Platform/import-repository/lib/api"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/controller"
	"github.com/SENERGY-Platform/import-repository/lib/database"
	"github.com/SENERGY-Platform/import-repository/lib/source/consumer"
	"github.com/SENERGY-Platform/import-repository/lib/source/consumer/listener"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func Start(conf config.Config, ctx context.Context, wg *sync.WaitGroup) (err error) {
	return StartWithPermv2Client(conf, ctx, wg, permV2.New(conf.PermissionsV2Url))
}

func StartWithPermv2Client(conf config.Config, ctx context.Context, wg *sync.WaitGroup, permV2Client permV2.Client) (err error) {
	db, err := database.New(conf, ctx, wg)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		return err
	}

	ctrl, err := controller.New(conf, db, permV2Client)
	if err != nil {
		log.Println("ERROR: unable to start control", err)
		return err
	}

	err = ctrl.Migrate()
	if err != nil {
		log.Println("ERROR: unable to migrate", err)
		return err
	}

	_, err = consumer.NewConsumer(ctx, wg, conf.KafkaBootstrap, []string{conf.UsersTopic}, conf.GroupId, consumer.Earliest,
		listener.UsersListenerFactory(ctrl), consumer.HandleError, conf.Debug)
	if err != nil {
		log.Println("WARNING: unable to start source, retrying periodically...", err)
	}

	err = api.Start(conf, ctrl)
	if err != nil {
		log.Println("ERROR: unable to start api", err)
		return err
	}

	return err
}
