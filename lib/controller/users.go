/*
 * Copyright 2021 InfAI (CC SES)
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

import "github.com/SENERGY-Platform/import-repository/lib/auth"

func (this *Controller) DeleteUser(userId string) error {
	token, err := auth.CreateToken("device-manager", userId)
	if err != nil {
		return err
	}
	//devices
	devicesToDelete, userToDeleteFromDevices, err := this.com.ResourcesEffectedByUserDelete(token, this.config.ImportTypeTopic)
	if err != nil {
		return err
	}
	for _, id := range devicesToDelete {
		err = this.producer.PublishImportTypeDelete(id, userId)
		if err != nil {
			return err
		}
	}
	for _, id := range userToDeleteFromDevices {
		err = this.producer.PublishDeleteUserRights(this.config.ImportTypeTopic, id, userId)
		if err != nil {
			return err
		}
	}

	return nil
}
