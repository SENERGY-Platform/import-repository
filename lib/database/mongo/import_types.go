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
	"regexp"
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

func (this *Mongo) ListImportTypes(ctx context.Context, listOptions model.ImportTypeListOptions) (result []model.ImportType, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = "name.asc"
	}

	sortby := listOptions.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")

	direction := int32(1)
	if strings.HasSuffix(listOptions.SortBy, ".desc") {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[idKey] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[nameKey] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	//TODO: SNRGY-3572 filter by listOptions.Criteria
	//	test with tests/list_test.go::TestList()

	cursor, err := this.importTypeCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, total, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return result, total, err
	}
	total, err = this.importTypeCollection().CountDocuments(ctx, filter)
	if err != nil {
		return result, total, err
	}
	return result, total, err
}

func (this *Mongo) SetImportType(ctx context.Context, importType model.ImportType) error {
	oldConfigs := []model.ImportConfig{}
	for idx, config := range importType.Configs {
		oldConfigs = append(oldConfigs, model.ImportConfig{
			Name:               config.Name,
			Description:        config.Description,
			Type:               config.Type,
			DefaultValue:       config.DefaultValue,
			DefaultValueString: config.DefaultValueString,
		})
		err := configToWrite(&config)
		if err != nil {
			return err
		}
		importType.Configs[idx] = config
	}
	_, err := this.importTypeCollection().ReplaceOne(ctx, bson.M{idKey: importType.Id}, importType, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	for i := range importType.Configs {
		importType.Configs[i] = oldConfigs[i]
	}
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
