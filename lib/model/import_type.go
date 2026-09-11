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

import "github.com/SENERGY-Platform/models/go/models"

type ImportType = models.ImportType

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
	for _, aspectId := range c.AspectIds {
		if aspectId == "" {
			continue
		}
		if c.FunctionId != "" {
			aspectFunctions[aspectId+"_"+c.FunctionId] = nil
		}
		aspects[aspectId] = nil
	}
	if c.FunctionId != "" {
		functions[c.FunctionId] = nil
	}
	for _, sub := range c.SubContentVariables {
		fillAspectFunctions(aspectFunctions, aspects, functions, sub)
	}
}

// ImportConfig declares a config of an import type. The shared model calls it
// ImportTypeConfig, to keep it apart from the config *values* an import instance carries.
type ImportConfig = models.ImportTypeConfig

type ImportTypeListOptions struct {
	Ids      []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search   string
	Limit    int64                      //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset   int64                      //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy   string                     //default name.asc
	Criteria []ImportTypeFilterCriteria //filter; ignored if nil

	//AndCombineCriteriaAspectIds interprets the AspectIds of a criteria as an AND: the same
	//content variable has to carry all of them, and each of them covers its aspect subtree.
	//The default is an OR, where the caller names the aspect subtree it wants itself.
	AndCombineCriteriaAspectIds bool
}

// ImportTypeFilterCriteria is the shared filter-criteria. Its deprecated AspectId is an alias
// for a single element of AspectIds and is folded into that list before the criteria is
// resolved, the same way ContentVariable.AspectId is folded on write.
type ImportTypeFilterCriteria = models.ImportTypeFilterCriteria

// ImportTypeQueryOptions is ImportTypeListOptions with the aspects of every filter-criteria
// resolved for the database. The controller resolves them, because an aspect subtree is known
// to the device-repository and not to this service.
type ImportTypeQueryOptions struct {
	Ids      []string //filter; ignored if nil; Ids == []string{} will return an empty list;
	Search   string
	Limit    int64
	Offset   int64
	SortBy   string                    //default name.asc
	Criteria []ImportTypeCriteriaQuery //filter; ignored if nil
}

// ImportTypeCriteriaQuery is a filter-criteria as the database evaluates it. A content
// variable matches a set of AspectIdSets if it carries any of its ids (an OR), and it has to
// match every set (an AND).
type ImportTypeCriteriaQuery struct {
	FunctionId   string
	AspectIdSets [][]string
}
