/*
 * Copyright 2023 InfAI (CC SES)
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

func (this *Controller) Republish() error {
	var limit int64 = 1000
	var offset int64 = 0

	for {
		ctx, cancel := getTimeoutContext()
		defer cancel()
		importTypes, err := this.db.ListImportTypes(ctx, limit, offset, "id.asc")
		if err != nil {
			return err
		}
		for _, t := range importTypes {
			err = this.producer.PublishImportType(t, t.Owner)
			if err != nil {
				return err
			}
		}
		if int64(len(importTypes)) < limit {
			break
		}
		offset += int64(len(importTypes))
	}
	return nil
}
