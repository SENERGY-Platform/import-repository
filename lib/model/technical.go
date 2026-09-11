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

// The types of the import-type model live in the shared model. They are aliased rather than
// imported directly, because the client of this service is compiled against these names.

type Type = models.Type

const (
	String  = models.String
	Integer = models.Integer
	Float   = models.Float
	Boolean = models.Boolean

	List      = models.List
	Structure = models.Structure
)

// ContentVariable is the output description of an import type. It carries AspectIds; the
// deprecated AspectId is folded into that list, see content_variable_aspects.go.
type ContentVariable = models.ImportContentVariable
