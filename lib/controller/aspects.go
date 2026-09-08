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

package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/cache"
)

const aspectNodeCacheExpiration = time.Minute

// getAspectNode loads an aspect node from the device-repository. The aspect hierarchy changes
// rarely while a filter-criteria may name the same aspect on every request, so the node is
// cached. An unknown aspect is tolerated: it is answered with a node that has no descendents
// and therefore covers only itself, the way the device-repository tolerates an unknown aspect
// in its own filter-criteria.
func (this *Controller) getAspectNode(id string) (result models.AspectNode, err error) {
	return cache.Use(this.aspectCache, "aspect-nodes."+id, func() (models.AspectNode, error) {
		node, err, code := this.deviceRepoClient.GetAspectNode(id)
		if code == http.StatusNotFound {
			log.Logger.Warn("aspect id of filter-criteria not found as aspect-node", "aspectId", id)
			return models.AspectNode{Id: id}, nil
		}
		if err != nil {
			return node, err
		}
		return node, nil
	}, func(node models.AspectNode) error {
		if node.Id == "" {
			return errors.New("invalid aspect-node loaded from cache")
		}
		return nil
	}, aspectNodeCacheExpiration)
}

// aspectIdWithDescendents returns the aspect subtree an aspect of a filter-criteria covers.
func (this *Controller) aspectIdWithDescendents(id string) (result []string, err error) {
	node, err := this.getAspectNode(id)
	if err != nil {
		return nil, err
	}
	return append([]string{id}, node.DescendentIds...), nil
}
