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

package mongo

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
	"sync"
	"time"
)

type Mongo struct {
	config config.Config
	client *mongo.Client
}

var CreateCollections = []func(db *Mongo) error{}

func New(conf config.Config, ctx context.Context, wg *sync.WaitGroup) (*Mongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoUrl))
	if err != nil {
		return nil, err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		_ = client.Disconnect(context.Background())
		wg.Done()
	}()
	db := &Mongo{config: conf, client: client}
	for _, creators := range CreateCollections {
		err = creators(db)
		if err != nil {
			_ = client.Disconnect(context.Background())
			return nil, err
		}
	}
	err = db.migrateImportTypeCriteria()
	if err != nil {
		return db, err
	}
	return db, nil
}

func (this *Mongo) CreateId() string {
	return uuid.NewString()
}

func (this *Mongo) ensureIndex(collection *mongo.Collection, indexname string, indexKey string, asc bool, unique bool) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{indexKey, direction}},
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	return err
}

func (this *Mongo) ensureCompoundIndex(collection *mongo.Collection, indexname string, asc bool, unique bool, indexKeys ...string) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	keys := []bson.E{}
	for _, key := range indexKeys {
		keys = append(keys, bson.E{Key: key, Value: direction})
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D(keys),
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	return err
}

func (this *Mongo) Disconnect() {
	log.Println(this.client.Disconnect(context.Background()))
}

func getBsonFieldName(obj interface{}, fieldName string) (bsonName string, err error) {
	field, found := reflect.TypeOf(obj).FieldByName(fieldName)
	if !found {
		return "", errors.New("field '" + fieldName + "' not found")
	}
	tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
	if err != nil {
		return "", err
	}
	return tags.Name, err
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
