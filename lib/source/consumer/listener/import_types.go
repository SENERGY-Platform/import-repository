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

package listener

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/import-repository/lib/source"
	"github.com/SENERGY-Platform/import-repository/lib/source/consumer"
	"log"
	"time"
)

type Controller interface {
	SetImportTypeInDB(device model.ImportType) error
	DeleteImportTypeFromDB(id string) error
	DeleteUser(id string) error
}

func ImportTypesListenerFactory(control Controller) func(topic string, msg []byte, time time.Time) error {
	return func(_ string, msg []byte, _ time.Time) (err error) {
		command := source.ImportTypeCommand{}
		err = json.Unmarshal(msg, &command)
		if err != nil {
			return
		}
		switch command.Command {
		case "PUT":
			return control.SetImportTypeInDB(model.ShrinkImportType(command.ImportType))
		case "DELETE":
			return control.DeleteImportTypeFromDB(command.Id)
		case "RIGHTS":
			return nil
		}
		return errors.New("unable to handle command: " + string(msg))
	}
}

func HandleError(err error, _ *consumer.Consumer) {
	log.Println(err)
	panic("Failing hard in order to prevent committing of invalid offsets!")
}
