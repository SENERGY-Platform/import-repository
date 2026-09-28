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

package mongo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func randomHex(t *testing.T) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// TestNewAuthenticates needs root credentials of a throwaway server with access control, given by
// MONGO_AUTH_TEST_URL, _USER and _PASSWORD; it creates and drops its own users and databases.
func TestNewAuthenticates(t *testing.T) {
	url, rootUser, rootPassword := os.Getenv("MONGO_AUTH_TEST_URL"), os.Getenv("MONGO_AUTH_TEST_USER"), os.Getenv("MONGO_AUTH_TEST_PASSWORD")
	if testing.Short() || url == "" || rootUser == "" || rootPassword == "" {
		t.Skip("needs MONGO_AUTH_TEST_URL, MONGO_AUTH_TEST_USER and MONGO_AUTH_TEST_PASSWORD, not in -short")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	root, err := mongo.Connect(ctx, options.Client().ApplyURI(url).SetAuth(options.Credential{Username: rootUser, Password: rootPassword, AuthSource: "admin"}))
	if err != nil {
		t.Fatal(err)
	}
	// Registered first so it runs after the user cleanups, which need the connection.
	t.Cleanup(func() { _ = root.Disconnect(context.Background()) })

	suffix := randomHex(t)
	database, otherDatabase := "import_repository_authtest_"+suffix, "other_authtest_"+suffix
	user, otherUser := "import-repository-"+suffix, "other-"+suffix
	password, otherPassword := randomHex(t), randomHex(t)
	createUser := func(name, pw, role, db string) {
		// Registered before creation and tolerant of a missing user, so a half-done setup is cleaned up too.
		t.Cleanup(func() {
			err := root.Database("admin").RunCommand(context.Background(), bson.D{{Key: "dropUser", Value: name}}).Err()
			var cmdErr mongo.CommandError
			if err != nil && !(errors.As(err, &cmdErr) && cmdErr.Code == 11) {
				t.Errorf("unable to drop user %v: %v", name, err)
			}
			if err := root.Database(db).Drop(context.Background()); err != nil {
				t.Errorf("unable to drop database %v: %v", db, err)
			}
		})
		err := root.Database("admin").RunCommand(ctx, bson.D{
			{Key: "createUser", Value: name},
			{Key: "pwd", Value: pw},
			{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: role}, {Key: "db", Value: db}}}},
		}).Err()
		if err != nil {
			t.Fatal(err)
		}
	}
	createUser(user, password, "readWrite", database)
	createUser(otherUser, otherPassword, "readWrite", otherDatabase)

	cases := []struct {
		name     string
		user     string
		password string
		wantErr  bool
	}{
		{"correct user", user, password, false},
		{"no credentials", "", "", true},
		{"user of another database", otherUser, otherPassword, true},
		{"wrong password", user, password + "-wrong", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := config.Config{
				MongoUrl:                  url,
				MongoUser:                 c.user,
				MongoPassword:             c.password,
				MongoAuthSource:           "admin",
				MongoDatabase:             database,
				MongoImportTypeCollection: "importtype",
			}
			svcCtx, stop := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			defer wg.Wait()
			defer stop()
			db, err := New(conf, svcCtx, wg)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if err != nil {
				for _, pw := range []string{password, otherPassword, rootPassword} {
					if strings.Contains(err.Error(), pw) {
						t.Fatal("error contains a password")
					}
				}
				if !strings.HasPrefix(err.Error(), "mongo startup check failed: ") {
					t.Errorf("unexpected error: %v", err)
				}
				t.Log(err)
				return
			}
			importType := model.ImportType{Id: "authtest", Name: "authtest", Output: model.ContentVariable{Name: "output"}}
			if err = db.SetImportType(context.Background(), importType); err != nil {
				t.Fatalf("write after successful startup failed: %v", err)
			}
			got, exists, err := db.GetImportType(context.Background(), importType.Id)
			if err != nil || !exists || got.Id != importType.Id {
				t.Errorf("read after successful startup: %v, %v, %v", got, exists, err)
			}
		})
	}
}
