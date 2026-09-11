/*
 * Copyright 2026 InfAI (CC SES)
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

import "slices"

//ContentVariable.AspectId is deprecated in favor of ContentVariable.AspectIds. It is an
//alias for a single element list and is folded into that list on write, so that everything
//behind the write evaluates AspectIds only. The deprecated field is kept in sync on read,
//so that clients which know only one of the two fields keep working.

// SetContentVariableAspectIdsOnWrite adds the deprecated AspectId to AspectIds, so that only
// AspectIds has to be interpreted afterwards. It is used by the controller on write and by
// the database migration, which rebuilds the derived criteria of stored import types.
func SetContentVariableAspectIdsOnWrite(importType *ImportType) {
	forEachContentVariable(&importType.Output, addContentVariableAspectIdToAspectIds)
}

// SetContentVariableAspectIdsOnRead makes both fields consistent. AspectIds is filled from
// the deprecated AspectId first, because import types stored before the migration carry only
// AspectId. AspectId is then set to the alphabetically first entry of AspectIds, to serve
// clients that do not know AspectIds yet.
func SetContentVariableAspectIdsOnRead(importType *ImportType) {
	forEachContentVariable(&importType.Output, syncContentVariableAspectIds)
}

func SetContentVariableAspectIdsOnReadList(importTypes []ImportType) {
	for i := range importTypes {
		SetContentVariableAspectIdsOnRead(&importTypes[i])
	}
}

// FilterCriteriaAspectIds returns the aspects of a filter-criteria with the deprecated
// AspectId folded into the list, so that only AspectIds has to be interpreted afterwards.
// The shared filter-criteria carries both fields; this service names only the list.
func FilterCriteriaAspectIds(criteria ImportTypeFilterCriteria) []string {
	if criteria.AspectId == "" || slices.Contains(criteria.AspectIds, criteria.AspectId) {
		return criteria.AspectIds
	}
	return append(slices.Clone(criteria.AspectIds), criteria.AspectId)
}

func syncContentVariableAspectIds(variable *ContentVariable) {
	addContentVariableAspectIdToAspectIds(variable)
	if len(variable.AspectIds) == 0 {
		return
	}
	sorted := slices.Clone(variable.AspectIds)
	slices.Sort(sorted)
	variable.AspectId = sorted[0]
}

func addContentVariableAspectIdToAspectIds(variable *ContentVariable) {
	if variable.AspectId != "" && !slices.Contains(variable.AspectIds, variable.AspectId) {
		variable.AspectIds = append(variable.AspectIds, variable.AspectId)
	}
}

func forEachContentVariable(variable *ContentVariable, f func(variable *ContentVariable)) {
	f(variable)
	for i := range variable.SubContentVariables {
		forEachContentVariable(&variable.SubContentVariables[i], f)
	}
}
