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
	// if the count of import types becomes to big, we may want to list for all permissions separately, filtered by the deleted user
	// eg:
	// readIds, err, _ := this.permV2Client.ListAccessibleResourceIds(tokenOfUser, ..., permV2.Read)
	// ids = append(ids, readIds...)
	// writeIds, err, _ := this.permV2Client.ListAccessibleResourceIds(tokenOfUser, ..., permV2.Write)
	// ids = append(ids, readIds...)
	// ...
	// importTypes, err, _ := this.permV2Client.ListResourcesWithAdminPermission(permV2.InternalAdminToken, PermV2Topic, permV2.ListOptions{Ids: ids})
	importTypes, err, _ := this.permV2Client.ListResourcesWithAdminPermission(permV2.InternalAdminToken, PermV2Topic, permV2.ListOptions{})
	if err != nil {
		return err
	}
	for _, importType := range importTypes {
		_, ok := importType.UserPermissions[userId]
		if !ok {
			continue // user has no rights to that import type
		}

		// remove user permissions
		delete(importType.UserPermissions, userId)

		//other admin exists?
		found := false
		for _, perm := range importType.UserPermissions {
			if perm.Administrate {
				found = true
				break
			}
		}

		//no other admin user
		if found {
			_, err, _ = this.permV2Client.SetPermission(permV2.InternalAdminToken, PermV2Topic, importType.Id, importType.ResourcePermissions)
			if err != nil {
				return err
			}
		} else {
			err, _ = this.deleteImportType(importType.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
