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
	config       config.Config
	syncProducer sarama.SyncProducer
}

func New(conf config.Config, ctx context.Context, wg *sync.WaitGroup) (*Producer, error) {
	p := &Producer{config: conf}
	var err error
	p.syncProducer, err = p.ensureConnection()
	wg.Add(1)
	go func() {
		<-ctx.Done()
		if p.syncProducer != nil {
			_ = p.syncProducer.Close()
		}
		wg.Done()
	}()
	return p, err
}

func (producer *Producer) ensureConnection() (syncProducer sarama.SyncProducer, err error) {
	if producer.syncProducer != nil {
		return producer.syncProducer, nil
	}
	kafkaConf := sarama.NewConfig()
	kafkaConf.Producer.Return.Successes = true
	syncP, err := sarama.NewSyncProducer([]string{producer.config.KafkaBootstrap}, kafkaConf)
	if err != nil {
		producer.syncProducer = syncP
	}
	return syncP, err
}
