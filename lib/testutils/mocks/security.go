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

package mocks

import (
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/auth"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"runtime/debug"
)

type Security struct {
	access map[string]bool
}

func NewSecurity() *Security {
	return &Security{access: map[string]bool{}}
}

func (this *Security) CheckBool(jwt auth.Token, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	return this.access[this.getKey(kind, id)], nil
}

func (this *Security) CheckMultiple(jwt auth.Token, kind string, ids []string, action model.AuthAction) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range ids {
		result[id], _ = this.CheckBool(jwt, kind, id, action)
	}
	return result, nil
}

func (this *Security) getKey(kind string, id string) string {
	return kind + "/" + id
}

func (this *Security) Set(kind string, id string, access bool) {
	this.access[this.getKey(kind, id)] = access
}

func (this *Security) GetAsUser(jwt auth.Token, url string, result *[]interface{}) (err error) {
	debug.PrintStack()
	return errors.New("not implemented")
}
