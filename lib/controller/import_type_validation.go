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
 *
 */
package controller

import (
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"math"
	"net/http"
)

func (this *Controller) ValidateImportType(token jwt.Token, importType model.ImportType) (err error, code int) {
	if len(importType.Name) == 0 {
		return errors.New("name might not be empty"), http.StatusBadRequest
	}

	if len(importType.Image) == 0 {
		return errors.New("image might not be empty"), http.StatusBadRequest
	}

	confNames := []string{}
	for _, conf := range importType.Configs {
		if contains(confNames, conf.Name) {
			return errors.New("duplicate config name"), http.StatusBadRequest
		}
		confNames = append(confNames, conf.Name)
		if !validateConfig(conf) {
			return errors.New("invalid config"), http.StatusBadRequest
		}
	}

	ok, err := this.validateContentVariable(token, importType.Output)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("invalid output"), http.StatusBadRequest
	}
	return
}

func validateConfig(conf model.ImportConfig) (valid bool) {
	valid = true
	if len(conf.Name) == 0 ||
		(conf.Type != model.String &&
			conf.Type != model.Integer &&
			conf.Type != model.Float &&
			conf.Type != model.List &&
			conf.Type != model.Structure &&
			conf.Type != model.Boolean) {
		return false
	}
	if conf.DefaultValue != nil {
		switch conf.Type {
		case model.String:
			_, valid = conf.DefaultValue.(string)
			break
		case model.Integer:
			val, validInner := conf.DefaultValue.(float64)
			valid = validInner && math.Mod(val, 1) == 0
			break
		case model.Float:
			_, valid = conf.DefaultValue.(float64)
			break
		case model.List:
			_, valid = conf.DefaultValue.([]interface{})
			break
		case model.Structure:
			_, valid = conf.DefaultValue.(map[string]interface{})
			break
		case model.Boolean:
			_, valid = conf.DefaultValue.(bool)
			break
		}
	}
	return valid
}

func (this *Controller) validateContentVariable(token jwt.Token, variable model.ContentVariable) (valid bool, err error) {
	valid, characteristicIds, functionIds, aspectIds := this.validateContentVariableStep(token, variable)
	if !valid {
		return false, nil
	}
	if len(characteristicIds) > 0 {
		for _, characteristicId := range characteristicIds {
			_, err, code := this.deviceRepoClient.GetCharacteristic(characteristicId)
			if err != nil || code > 299 {
				return false, err
			}
		}
	}
	if len(functionIds) > 0 {
		for _, functionId := range functionIds {
			_, err, code := this.deviceRepoClient.GetFunction(functionId)
			if err != nil || code > 299 {
				return false, err
			}
		}
	}
	if len(aspectIds) > 0 {
		uniqueIds := []string{}
		for _, id := range aspectIds {
			if !contains(uniqueIds, id) {
				uniqueIds = append(uniqueIds, id)
			}
		}

		for _, aspectId := range uniqueIds {
			_, err, code := this.deviceRepoClient.GetAspectNode(aspectId)
			if err != nil || code > 299 {
				return false, err
			}
		}
	}
	return valid, err
}

func (this *Controller) validateContentVariableStep(token jwt.Token, variable model.ContentVariable) (valid bool, characteristicIds []string, functionIds []string, aspectIds []string) {
	if len(variable.Name) == 0 || len(variable.Type) == 0 {
		return false, characteristicIds, functionIds, aspectIds
	}
	if variable.Type != model.String &&
		variable.Type != model.Integer &&
		variable.Type != model.Float &&
		variable.Type != model.List &&
		variable.Type != model.Structure &&
		variable.Type != model.Boolean {
		return false, characteristicIds, functionIds, aspectIds
	}
	if variable.Type != model.Structure && variable.Type != model.List && len(variable.SubContentVariables) > 0 {
		return false, characteristicIds, functionIds, aspectIds
	}
	if len(variable.CharacteristicId) > 0 {
		characteristicIds = append(characteristicIds, variable.CharacteristicId)
	}
	if len(variable.FunctionId) > 0 {
		functionIds = append(functionIds, variable.FunctionId)
	}
	if len(variable.AspectId) > 0 {
		aspectIds = append(aspectIds, variable.AspectId)
	}
	for _, subVariable := range variable.SubContentVariables {
		validInner, subCharacteristicIds, subFunctionIds, subAspectIds := this.validateContentVariableStep(token, subVariable)
		if !validInner {
			return validInner, characteristicIds, functionIds, aspectIds
		}
		characteristicIds = append(characteristicIds, subCharacteristicIds...)
		functionIds = append(functionIds, subFunctionIds...)
		aspectIds = append(aspectIds, subAspectIds...)
	}
	return true, characteristicIds, functionIds, aspectIds
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
