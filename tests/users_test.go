/*
 * Copyright 2021 InfAI (CC SES)
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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/controller"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/import-repository/lib/source/consumer/listener"
	permV2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func TestUserDelete(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	permv2Client, conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	user1, err := createToken("test", "user1")
	if err != nil {
		t.Error(err)
		return
	}
	user2, err := createToken("test", "user2")
	if err != nil {
		t.Error(err)
		return
	}

	conf.Debug = true

	ids := []string{}
	t.Run("create import-types", initImportTypes(conf, user1, user2, &ids))

	time.Sleep(30 * time.Second)

	t.Run("change permissions", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			id := ids[i]
			err = setPermission(permv2Client, id, permV2.ResourcePermissions{
				UserPermissions: map[string]permV2.PermissionsMap{
					user2.Sub: permV2.PermissionsMap{
						Read:         true,
						Write:        true,
						Execute:      true,
						Administrate: true,
					},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}

		for i := 2; i < 4; i++ {
			id := ids[i]
			err = setPermission(permv2Client, id, permV2.ResourcePermissions{
				UserPermissions: map[string]permV2.PermissionsMap{
					user2.Sub: permV2.PermissionsMap{
						Read:         true,
						Write:        true,
						Execute:      true,
						Administrate: true,
					},
					user1.Sub: permV2.PermissionsMap{
						Read:    true,
						Execute: true,
					},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}

		for i := 10; i < 12; i++ {
			id := ids[i]
			err = setPermission(permv2Client, id, permV2.ResourcePermissions{
				UserPermissions: map[string]permV2.PermissionsMap{
					user1.Sub: permV2.PermissionsMap{
						Read:         true,
						Write:        true,
						Execute:      true,
						Administrate: true,
					},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}

		for i := 12; i < 14; i++ {
			id := ids[i]
			err = setPermission(permv2Client, id, permV2.ResourcePermissions{
				UserPermissions: map[string]permV2.PermissionsMap{
					user1.Sub: permV2.PermissionsMap{
						Read:         true,
						Write:        true,
						Execute:      true,
						Administrate: true,
					},
					user2.Sub: permV2.PermissionsMap{
						Read:    true,
						Execute: true,
					},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}

		// 10, 11, 12, 13, 14 for user1 rwxa
		for i := 13; i < 15; i++ {
			id := ids[i]
			err = setPermission(permv2Client, id, permV2.ResourcePermissions{
				UserPermissions: map[string]permV2.PermissionsMap{
					user1.Sub: permV2.PermissionsMap{
						Read:         true,
						Write:        true,
						Execute:      true,
						Administrate: true,
					},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	time.Sleep(30 * time.Second)

	t.Run("check user1 before delete", checkUserImportTypes(conf, user1, ids[:15]))
	t.Run("check user2 before delete", checkUserImportTypes(conf, user2, append(append([]string{}, ids[:4]...), ids[10:]...)))

	t.Run("delete user1", func(t *testing.T) {
		kafkaConf := sarama.NewConfig()
		kafkaConf.Producer.Return.Successes = true
		syncP, err := sarama.NewSyncProducer([]string{conf.KafkaBootstrap}, kafkaConf)
		if err != nil {
			t.Error(err)
			return
		}
		defer syncP.Close()
		cmd := listener.UserCommandMsg{
			Command: "DELETE",
			Id:      user1.GetUserId(),
		}
		message, err := json.Marshal(cmd)
		if err != nil {
			t.Error(err)
			return
		}
		_, _, err = syncP.SendMessage(
			&sarama.ProducerMessage{
				Topic: conf.UsersTopic,
				Value: sarama.StringEncoder(message),
				Key:   sarama.StringEncoder(cmd.Id),
			})
		if err != nil {
			t.Error(err)
			return
		}
	})

	time.Sleep(5 * time.Second)

	t.Run("check user1 after delete", checkUserImportTypes(conf, user1, []string{}))
	t.Run("check user2 after delete", checkUserImportTypes(conf, user2, append(append(append([]string{}, ids[:4]...), ids[10:12]...), ids[14:]...)))

}

type IdWrapper struct {
	Id string `json:"id"`
}

func initImportTypes(config config.Config, user1 jwt.Token, user2 jwt.Token, createdIds *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		for i := 0; i < 10; i++ {
			temp := IdWrapper{}
			err := jwtpostjson(user1,
				"http://localhost:"+config.ServerPort+"/import-types",
				model.ImportType{
					Name: strconv.Itoa(i),
					Configs: []model.ImportConfig{
						{
							Name: "struct",
							Type: model.Structure,
							DefaultValue: map[string]interface{}{
								"1234": "5678",
								"abc":  123,
								"def":  false,
							},
						},
					},
				}, &temp)
			if err != nil {
				t.Error(err)
				return
			}
			*createdIds = append(*createdIds, temp.Id)
		}
		for i := 10; i < 20; i++ {
			temp := IdWrapper{}
			err := jwtpostjson(user2,
				"http://localhost:"+config.ServerPort+"/import-types",
				model.ImportType{
					Name: strconv.Itoa(i),
					Configs: []model.ImportConfig{
						{
							Name: "struct",
							Type: model.Structure,
							DefaultValue: map[string]interface{}{
								"1234": "5678",
								"abc":  123,
								"def":  false,
							},
						},
					},
				}, &temp)
			if err != nil {
				t.Error(err)
				return
			}
			*createdIds = append(*createdIds, temp.Id)
		}
	}
}

func setPermission(permv2Client permV2.Client, id string, permissions permV2.ResourcePermissions) error {
	_, err, _ := permv2Client.SetPermission(permV2.InternalAdminToken, controller.PermV2Topic, id, permissions)
	return err
}

func checkUserImportTypes(conf config.Config, token jwt.Token, expectedIdsOrig []string) func(t *testing.T) {
	return func(t *testing.T) {
		expectedIds := []string{}
		temp, err := json.Marshal(expectedIdsOrig)
		if err != nil {
			t.Error(err)
			return
		}
		err = json.Unmarshal(temp, &expectedIds)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("GET", conf.PermissionsUrl+"/v3/resources/"+conf.ImportTypeTopic+"?rights=r&limit=100", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", token.Token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			resp.Body.Close()
			log.Println("DEBUG:", buf.String())
			err = errors.New("access denied")
			t.Error(err)
			return
		}

		devices := []map[string]interface{}{}
		err = json.NewDecoder(resp.Body).Decode(&devices)
		if err != nil {
			t.Error(err)
			return
		}
		actualIds := []string{}
		for _, device := range devices {
			id, ok := device["id"].(string)
			if !ok {
				t.Error("expect device id to be string", device)
				return
			}
			actualIds = append(actualIds, id)
		}
		sort.Strings(actualIds)
		sort.Strings(expectedIds)

		if !reflect.DeepEqual(actualIds, expectedIds) {
			t.Error(actualIds,
				"\n",
				expectedIds)
			return
		}
	}
}
