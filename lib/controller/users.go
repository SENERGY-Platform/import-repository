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

import (
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func (this *Controller) DeleteUser(userId string) error {
	importTypes, err, _ := this.permV2Client.ListResourcesWithAdminPermission(permV2.InternalAdminToken, PermV2Topic, permV2.ListOptions{})
	if err != nil {
		return err
	}
	for _, importType := range importTypes {
		perm, ok := importType.UserPermissions[userId]
		if !ok {
			continue // user has no rights to that import type
		}
		if !perm.Administrate {
			continue // user has no administrate rights to that import type
		}
		delete(importType.UserPermissions, userId) // find any user beside this one
		found := false
		for _, perm := range importType.UserPermissions {
			if perm.Administrate {
				found = true
			}
		}
		if !found {
			ctx, _ := getTimeoutContext()
			err = this.db.RemoveImportType(ctx, importType.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
