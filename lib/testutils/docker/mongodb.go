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

package docker

import (
	"context"
	"sync"

	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func MongoDB(ctx context.Context, wg *sync.WaitGroup) (hostport string, containerip string, err error) {
	log.Logger.Info("start mongo")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo:4.1.11",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForLog("waiting for connections"),
				wait.ForListeningPort("27017/tcp"),
			),
			Tmpfs: map[string]string{"/data/db": "rw"},
		},
		Started: true,
	})
	if err != nil {
		return "", "", err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := c.Terminate(context.Background())
		if err != nil {
			log.Logger.Debug("remove container mongo failed", attributes.ErrorKey, err)
		} else {
			log.Logger.Debug("remove container mongo")
		}
	}()

	containerip, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "27017/tcp")
	if err != nil {
		return "", "", err
	}
	hostport = temp.Port()

	return hostport, containerip, err
}
