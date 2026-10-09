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

import "errors"

// ErrUsageUnavailable means the usage of an import type could not be determined. Its text reaches
// the client, so wrap it only with static strings.
var ErrUsageUnavailable = errors.New("could not check whether the import type is used, nothing was deleted")

// ReleaseRef names a smart service release the caller may read.
type ReleaseRef struct {
	Id       string `json:"id"`
	DesignId string `json:"design_id"`
	Name     string `json:"name"`
}

// ImportRef names an import instance the caller may read.
type ImportRef struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// SmartServiceUsage is the smart-service-repository's answer: Releases and Instances count over
// all users, Readable lists only the releases the caller may read.
type SmartServiceUsage struct {
	Releases  int64        `json:"releases"`
	Instances int64        `json:"instances"`
	Readable  []ReleaseRef `json:"readable"`
}

// ImportUsage is import-deploy's answer: Instances counts over all users, Readable lists only the
// instances the caller may read.
type ImportUsage struct {
	Instances int64       `json:"instances"`
	Readable  []ImportRef `json:"readable"`
}

// DeleteOptions carries what a delete needs besides the id. Force deletes an import type that is
// still used; only administrators may set it.
type DeleteOptions struct {
	Force bool
}

// InUseError refuses a deletion because releases or import instances still use the import type.
// The text is static; the handler builds the response body from Usage.
type InUseError struct {
	Usage ImportTypeUsage
}

func (e *InUseError) Error() string {
	return "import type is still used by smart services or import instances"
}

// ImportTypeUsage is what uses an import type, from both sources.
type ImportTypeUsage struct {
	SmartServices SmartServiceUsage
	Imports       ImportUsage
}

// Used tells whether anything uses the import type.
func (u ImportTypeUsage) Used() bool {
	return u.SmartServices.Releases > 0 || u.Imports.Instances > 0
}

// InUseResponse is the 409 body of the delete.
type InUseResponse struct {
	Error            string       `json:"error"`
	Releases         int64        `json:"releases"`
	ReleaseInstances int64        `json:"release_instances"`
	ImportInstances  int64        `json:"import_instances"`
	ReadableReleases []ReleaseRef `json:"readable_releases"`
	ReadableImports  []ImportRef  `json:"readable_imports"`
}

func (e *InUseError) Response() InUseResponse {
	response := InUseResponse{
		Error:            e.Error(),
		Releases:         e.Usage.SmartServices.Releases,
		ReleaseInstances: e.Usage.SmartServices.Instances,
		ImportInstances:  e.Usage.Imports.Instances,
		ReadableReleases: e.Usage.SmartServices.Readable,
		ReadableImports:  e.Usage.Imports.Readable,
	}
	if response.ReadableReleases == nil {
		response.ReadableReleases = []ReleaseRef{}
	}
	if response.ReadableImports == nil {
		response.ReadableImports = []ImportRef{}
	}
	return response
}
