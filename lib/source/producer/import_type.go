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
	"encoding/json"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/import-repository/lib/source"
	"github.com/Shopify/sarama"
	"log"
	"runtime/debug"
)

func (this *Producer) PublishImportType(importType model.ImportType, userId string) error {
	cmd := source.ImportTypeCommand{Command: "PUT", Id: importType.Id, ImportType: model.ExtendImportType(importType), Owner: userId}
	return this.PublishImportTypeCommand(cmd)
}

func (this *Producer) PublishImportTypeDelete(id string, userId string) error {
	cmd := source.ImportTypeCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishImportTypeCommand(cmd)
}

func (this *Producer) PublishImportTypeCommand(cmd source.ImportTypeCommand) error {
	if this.config.LogLevel == "DEBUG" {
		log.Println("DEBUG: produce device", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	_, _, err = this.importTypes.SendMessage(&sarama.ProducerMessage{Topic: this.config.ImportTypeTopic, Value: sarama.StringEncoder(message), Key: sarama.StringEncoder(cmd.Id)})
	if err != nil {
		debug.PrintStack()
	}
	return err
}
