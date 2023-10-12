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
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
)

const idFieldName = "Id"
const nameFieldName = "Name"

var idKey string
var nameKey string

func init() {
	var err error
	idKey, err = getBsonFieldName(model.ImportType{}, idFieldName)
	if err != nil {
		log.Fatal(err)
	}
	nameKey, err = getBsonFieldName(model.ImportType{}, nameFieldName)
	if err != nil {
		log.Fatal(err)
	}

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoImportTypeCollection)
		err = db.ensureIndex(collection, "importTypeIdindex", idKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) importTypeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoImportTypeCollection)
}

func (this *Mongo) GetImportType(ctx context.Context, id string) (importType model.ImportType, exists bool, err error) {
	result := this.importTypeCollection().FindOne(ctx, bson.M{idKey: id})
	err = result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return importType, false, nil
		}
		return
	}
	err = result.Decode(&importType)
	for idx, config := range importType.Configs {
		err = configToRead(&config)
		if err != nil {
			return importType, true, err
		}
		importType.Configs[idx] = config
	}
	return importType, true, err
}

func (this *Mongo) ListImportTypes(ctx context.Context, limit int64, offset int64, sort string) (result []model.ImportType, err error) {
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := idKey
	switch parts[0] {
	case "id":
		sortby = idKey
	case "name":
		sortby = nameKey
	default:
		sortby = idKey
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	cursor, err := this.importTypeCollection().Find(ctx, bson.M{}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		importType := model.ImportType{}
		err = cursor.Decode(&importType)
		if err != nil {
			return nil, err
		}
		for idx, config := range importType.Configs {
			err = configToRead(&config)
			if err != nil {
				return result, err
			}
			importType.Configs[idx] = config
		}
		result = append(result, importType)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) SetImportType(ctx context.Context, importType model.ImportType) error {
	for idx, config := range importType.Configs {
		err := configToWrite(&config)
		if err != nil {
			return err
		}
		importType.Configs[idx] = config
	}
	_, err := this.importTypeCollection().ReplaceOne(ctx, bson.M{idKey: importType.Id}, importType, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveImportType(ctx context.Context, id string) error {
	_, err := this.importTypeCollection().DeleteOne(ctx, bson.M{idKey: id})
	return err
}

func configToWrite(config *model.ImportConfig) error {
	if config == nil {
		return errors.New("nil config")
	}
	if config.Type != model.Structure {
		return nil
	}
	bs, err := json.Marshal(config.DefaultValue)
	if err != nil {
		return err
	}
	s := string(bs)
	config.DefaultValueString = &s
	config.DefaultValue = nil
	return nil
}

func configToRead(config *model.ImportConfig) error {
	if config == nil {
		return errors.New("nil config")
	}
	if config.Type != model.Structure {
		return nil
	}
	if config.DefaultValueString == nil {
		return errors.New("nil DefaultValueString")
	}
	config.DefaultValue = map[string]interface{}{}
	err := json.Unmarshal([]byte(*config.DefaultValueString), &config.DefaultValue)
	if err != nil {
		return err
	}
	config.DefaultValueString = nil
	return nil
}
