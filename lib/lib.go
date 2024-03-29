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
	"github.com/SENERGY-Platform/import-repository/lib/api"
	"github.com/SENERGY-Platform/import-repository/lib/com"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/controller"
	"github.com/SENERGY-Platform/import-repository/lib/database"
	"github.com/SENERGY-Platform/import-repository/lib/source/consumer"
	"github.com/SENERGY-Platform/import-repository/lib/source/consumer/listener"
	"github.com/SENERGY-Platform/import-repository/lib/source/producer"
	"log"
	"sync"
)

func Start(conf config.Config, ctx context.Context, wg *sync.WaitGroup) (err error) {
	db, err := database.New(conf, ctx, wg)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		return err
	}

	perm, err := com.NewSecurity(conf)
	if err != nil {
		log.Println("ERROR: unable to create permission handler", err)
		return err
	}

	p, err := producer.New(conf, ctx, wg)
	if err != nil {
		log.Println("WARNING: producer unable to connect to kafka, retrying when needed...", err)
	}

	ctrl, err := controller.New(conf, db, perm, p)
	if err != nil {
		log.Println("ERROR: unable to start control", err)
		return err
	}

	if conf.RepublishStartup {
		err = ctrl.Republish()
		if err != nil {
			log.Println("ERROR: unable to republish control", err)
			return err
		}
	}

	_, err = consumer.NewConsumer(ctx, wg, conf.KafkaBootstrap, []string{conf.ImportTypeTopic}, conf.GroupId, consumer.Earliest,
		listener.ImportTypesListenerFactory(ctrl), listener.HandleError, conf.Debug)
	if err != nil {
		log.Println("WARNING: unable to start source, retrying periodically...", err)
	}

	_, err = consumer.NewConsumer(ctx, wg, conf.KafkaBootstrap, []string{conf.UsersTopic}, conf.GroupId, consumer.Earliest,
		listener.UsersListenerFactory(ctrl), listener.HandleError, conf.Debug)
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
