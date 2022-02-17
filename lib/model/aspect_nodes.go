/*
 * Copyright 2022 InfAI (CC SES)
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

package model

type AspectNode struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	RootId        string   `json:"root_id"`
	ParentId      string   `json:"parent_id"`
	ChildIds      []string `json:"child_ids"`
	AncestorIds   []string `json:"ancestor_ids"`
	DescendentIds []string `json:"descendent_ids"`
}
