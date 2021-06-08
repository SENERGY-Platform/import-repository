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

package producer

import (
	"context"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/Shopify/sarama"
	"sync"
)

type Producer struct {
	config      config.Config
	importTypes sarama.SyncProducer
}

func New(conf config.Config, ctx context.Context, wg *sync.WaitGroup) (*Producer, error) {
	kafkaConf := sarama.NewConfig()
	kafkaConf.Producer.Return.Successes = true
	importTypes, err := sarama.NewSyncProducer([]string{conf.KafkaBootstrap}, kafkaConf)
	if err != nil {
		return nil, err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		_ = importTypes.Close()
		wg.Done()
	}()
	return &Producer{config: conf, importTypes: importTypes}, nil
}
