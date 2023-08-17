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

package model

type ImportType struct {
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Image          string          `json:"image"`
	DefaultRestart bool            `json:"default_restart"`
	Configs        []ImportConfig  `json:"configs"`
	Output         ContentVariable `json:"output"`
	Owner          string          `json:"owner"`
	Cost           uint64          `json:"cost"`
}

type ImportTypeExtended struct {
	Id                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Image              string          `json:"image"`
	DefaultRestart     bool            `json:"default_restart"`
	Configs            []ImportConfig  `json:"configs"`
	ContentAspectIds   []string        `json:"content_aspect_ids"`
	ContentFunctionIds []string        `json:"content_function_ids"`
	Output             ContentVariable `json:"output"`
	AspectFunctions    []string        `json:"aspect_functions"`
	Owner              string          `json:"owner"`
	Cost               uint64          `json:"cost"`
}

func ExtendImportType(importType ImportType) ImportTypeExtended {
	ex := ImportTypeExtended{
		Id:             importType.Id,
		Name:           importType.Name,
		Description:    importType.Description,
		Image:          importType.Image,
		DefaultRestart: importType.DefaultRestart,
		Configs:        importType.Configs,
		Output:         importType.Output,
		Owner:          importType.Owner,
		Cost:           importType.Cost,
	}
	aspectFunctions := make(map[string]interface{})
	aspects := make(map[string]interface{})
	functions := make(map[string]interface{})
	fillAspectFunctions(aspectFunctions, aspects, functions, importType.Output)
	for k := range aspectFunctions {
		ex.AspectFunctions = append(ex.AspectFunctions, k)
	}
	for k := range aspects {
		ex.ContentAspectIds = append(ex.ContentAspectIds, k)
	}
	for k := range functions {
		ex.ContentFunctionIds = append(ex.ContentFunctionIds, k)
	}
	return ex
}

func ShrinkImportType(importType ImportTypeExtended) ImportType {
	return ImportType{
		Id:             importType.Id,
		Name:           importType.Name,
		Description:    importType.Description,
		Image:          importType.Image,
		DefaultRestart: importType.DefaultRestart,
		Configs:        importType.Configs,
		Output:         importType.Output,
		Owner:          importType.Owner,
		Cost:           importType.Cost,
	}
}

func fillAspectFunctions(aspectFunctions map[string]interface{}, aspects map[string]interface{}, functions map[string]interface{}, c ContentVariable) {
	if c.AspectId != "" && c.FunctionId != "" {
		aspectFunctions[c.AspectId+"_"+c.FunctionId] = nil
	}
	if c.AspectId != "" {
		aspects[c.AspectId] = nil
	}
	if c.FunctionId != "" {
		functions[c.FunctionId] = nil
	}
	for _, sub := range c.SubContentVariables {
		fillAspectFunctions(aspectFunctions, aspects, functions, sub)
	}
}

type ImportConfig struct {
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	Type               Type        `json:"type"`
	DefaultValue       interface{} `json:"default_value"`
	DefaultValueString *string     `json:"-"`
}
