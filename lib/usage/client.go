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

// Package usage asks the smart-service-repository and import-deploy what uses an import type.
package usage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SENERGY-Platform/import-repository/lib/model"
)

const (
	requestTimeout = 10 * time.Second
	// The answers list names the caller may read; this only bounds a misbehaving server.
	maxBodyBytes = 4 << 20
)

type Client struct {
	smartServiceRepoUrl string
	importDeployUrl     string
	http                *http.Client
}

func New(smartServiceRepoUrl string, importDeployUrl string) *Client {
	return &Client{
		smartServiceRepoUrl: strings.TrimRight(smartServiceRepoUrl, "/"),
		importDeployUrl:     strings.TrimRight(importDeployUrl, "/"),
		http:                &http.Client{Timeout: requestTimeout},
	}
}

// SmartServices returns the smart-service-repository's answer for the import type, asked with the
// caller's own token. Any failure, including a non-200 status, is an error: an unknown answer is
// not "unused".
func (c *Client) SmartServices(ctx context.Context, importTypeId string, auth string) (usage model.SmartServiceUsage, err error) {
	// Pointers, so that a 200 without a count is an error and not "zero".
	var answer struct {
		Releases  *int64             `json:"releases"`
		Instances *int64             `json:"instances"`
		Readable  []model.ReleaseRef `json:"readable"`
	}
	if err = c.get(ctx, c.smartServiceRepoUrl+"/resource-usage/import-types/"+url.PathEscape(importTypeId), auth, &answer); err != nil {
		return usage, err
	}
	if answer.Releases == nil || answer.Instances == nil {
		return usage, errors.New("smart service repository answer has no release or instance count")
	}
	usage.Releases = *answer.Releases
	usage.Instances = *answer.Instances
	usage.Readable = answer.Readable
	if usage.Readable == nil {
		usage.Readable = []model.ReleaseRef{}
	}
	return usage, nil
}

// Imports returns import-deploy's answer for the import type, asked like SmartServices.
func (c *Client) Imports(ctx context.Context, importTypeId string, auth string) (usage model.ImportUsage, err error) {
	var answer struct {
		Instances *int64            `json:"instances"`
		Readable  []model.ImportRef `json:"readable"`
	}
	if err = c.get(ctx, c.importDeployUrl+"/import-type-usage/"+url.PathEscape(importTypeId), auth, &answer); err != nil {
		return usage, err
	}
	if answer.Instances == nil {
		return usage, errors.New("import-deploy answer has no instance count")
	}
	usage.Instances = *answer.Instances
	usage.Readable = answer.Readable
	if usage.Readable == nil {
		usage.Readable = []model.ImportRef{}
	}
	return usage, nil
}

func (c *Client) get(ctx context.Context, address string, auth string, answer any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %d", req.URL.Host, resp.StatusCode)
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(answer); err != nil {
		return fmt.Errorf("decode answer of %s: %w", req.URL.Host, err)
	}
	return nil
}
