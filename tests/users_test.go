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
	"context"
	"encoding/json"
	"reflect"
	"slices"
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

	t.Run("check user1 before permission change", checkUserImportTypes(permv2Client, user1, ids[:10], ids))
	t.Run("check user2 before permission change", checkUserImportTypes(permv2Client, user2, ids[10:], ids))

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

		for i := 8; i < 12; i++ {
			id := ids[i]
			err = setPermission(permv2Client, id, permV2.ResourcePermissions{
				UserPermissions: map[string]permV2.PermissionsMap{
					user1.Sub: {
						Read:         true,
						Write:        true,
						Execute:      true,
						Administrate: true,
					},
					user2.Sub: {
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

		for i := 16; i < 18; i++ {
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

		for i := 18; i < 20; i++ {
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

	users1Expected := []string{}
	users1Expected = append(users1Expected, ids[2:12]...)
	users1Expected = append(users1Expected, ids[16:]...)
	t.Run("check user1 before delete", checkUserImportTypes(permv2Client, user1, users1Expected, ids))
	users2Expected := []string{}
	users2Expected = append(users2Expected, ids[:4]...)
	users2Expected = append(users2Expected, ids[8:18]...)
	t.Run("check user2 before delete", checkUserImportTypes(permv2Client, user2, users2Expected, ids))

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

	users2Expected = []string{}
	users2Expected = append(users2Expected, ids[:4]...)
	users2Expected = append(users2Expected, ids[8:16]...)
	t.Run("check user1 after delete", checkUserImportTypes(permv2Client, user1, []string{}, ids))
	t.Run("check user2 after delete", checkUserImportTypes(permv2Client, user2, users2Expected, ids))

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

func checkUserImportTypes(permV2Client permV2.Client, token jwt.Token, expectedIdsOrig []string, allIds []string) func(t *testing.T) {
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

		actualIds, err, _ := permV2Client.ListAccessibleResourceIds(token.Jwt(), controller.PermV2Topic, permV2.ListOptions{}, permV2.Read)
		if err != nil {
			t.Error(err)
			return
		}
		if actualIds == nil {
			actualIds = []string{}
		}
		sort.Strings(actualIds)
		sort.Strings(expectedIds)
		if !reflect.DeepEqual(actualIds, expectedIds) {
			aIndexes := listToIndexList(actualIds, allIds)
			eIndexes := listToIndexList(expectedIds, allIds)
			sort.Ints(aIndexes)
			sort.Ints(eIndexes)
			t.Error("\na=", aIndexes, "\ne=", eIndexes)
			return
		}
	}
}

func listToIndexList(list []string, allIds []string) (indexes []int) {
	for _, id := range list {
		indexes = append(indexes, slices.Index(allIds, id))
	}
	return indexes
}
