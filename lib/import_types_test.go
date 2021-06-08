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

package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/model"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

const name = "import-type-name"

func TestImportTypesIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg, conf, err := createTestEnv(ctx)
	defer func() {
		cancel()
		if wg != nil {
			wg.Wait()
		}
	}()

	if err != nil {
		t.Error(err)
		return
	}

	it := model.ImportType{
		Name: name,
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
	}
	it, err = createImportType(conf, it)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(500 * time.Millisecond)

	t.Run("not existing", func(t *testing.T) {
		testImportTypeReadNotFound(t, conf, "foobar")
	})
	t.Run("invalid creation", func(t *testing.T) {
		it2 := model.ImportType{
			Name: name,
			Id:   "someid",
		}
		_, err = createImportType(conf, it2)
		if err == nil {
			t.Error("could create import type with id")
		}
		it2 = model.ImportType{
			Name:  name,
			Owner: "someone",
		}
		_, err = createImportType(conf, it2)
		if err == nil {
			t.Error("could create import type with owner")
		}
	})
	t.Run("testImportTypeRead", func(t *testing.T) {
		testImportTypeRead(t, conf, it)
	})
	t.Run("testImportTypeUpdate", func(t *testing.T) {
		it.Name = "new-name"
		err = updateImportType(conf, it, it.Id)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(500 * time.Millisecond)
		testImportTypeRead(t, conf, it)
	})
	t.Run("testImportTypeInvalidUpdate", func(t *testing.T) {
		err = updateImportType(conf, it, it.Id+"1")
		if err == nil {
			t.Error("could update with mismatching ids")
			return
		}
		it.Owner = "new-owner"
		err = updateImportType(conf, it, it.Id+"1")
		if err == nil {
			t.Error("could update owner")
			return
		}
		it.Owner = ""
		oldId := it.Id
		it.Id = "newID"
		err = updateImportType(conf, it, it.Id)
		if err == nil {
			t.Error("could update with own id")
			return
		}
		it.Id = oldId
	})
	t.Run("testImportTypeReadNotAllowed", func(t *testing.T) {
		testImportTypeReadNotAllowed(t, conf, it.Id)
	})
	t.Run("testImportTypeDelete", func(t *testing.T) {
		testImportTypeDelete(t, conf, it.Id)
		time.Sleep(500 * time.Millisecond)
		testImportTypeReadNotFound(t, conf, it.Id)
	})
}

func createImportType(conf config.Config, it model.ImportType) (resp model.ImportType, err error) {
	endpoint := "http://localhost:" + conf.ServerPort + "/import-types"
	err = userjwt.PostJSON(endpoint, it, &resp)
	return
}

func updateImportType(conf config.Config, it model.ImportType, id string) (err error) {
	endpoint := "http://localhost:" + conf.ServerPort + "/import-types/" + url.PathEscape(id)
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(it)
	if err != nil {
		return err
	}
	resp, err := jwtput(userjwt, endpoint, "application/json", b)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("wrong status code")
	}
	return
}

func testImportTypeReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/import-types/" + url.PathEscape(id)
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}

func testImportTypeRead(t *testing.T, conf config.Config, expectedImportTypes ...model.ImportType) {
	for _, expected := range expectedImportTypes {
		endpoint := "http://localhost:" + conf.ServerPort + "/import-types/" + url.PathEscape(expected.Id)
		resp, err := userjwt.Get(endpoint)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := ioutil.ReadAll(resp.Body)
			t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
			return
		}

		result := model.ImportType{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(expected, result) {
			t.Error("unexpected result", expected, result)
			return
		}
	}

}

func testImportTypeReadNotAllowed(t *testing.T, conf config.Config, id string) {
	resp, err := userjwt2.Get("http://localhost:" + conf.ServerPort + "/import-types/" + url.PathEscape(id))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Error("unexpected status code", http.StatusForbidden, resp.StatusCode)
		return
	}
	return
}

func testImportTypeDelete(t *testing.T, conf config.Config, id string) {
	resp, err := jwtdelete(userjwt, "http://localhost:"+conf.ServerPort+"/import-types/"+url.PathEscape(id))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Error("unexpected status code", http.StatusNoContent, resp.StatusCode)
		return
	}
	return
}
