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

/*
this file contains code needed to create the test environment
*/

package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/import-repository/lib/auth"
	"github.com/SENERGY-Platform/import-repository/lib/config"
	"github.com/SENERGY-Platform/import-repository/lib/testutils/docker"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

var userjwt auth.Token
var userjwt2 auth.Token

func init() {
	var err error

	userjwt, err = auth.ParseAuthToken("Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ")
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
	}
	userjwt2, err = auth.ParseAuthToken("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyMzMzMzMzfQ.DYBskZCLd-xyDqYkyesX-jBhwPJbHDoLhc83Q2H_bGM")
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
	}
}

//const userid = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"

func jwtdelete(jwt auth.Token, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", jwt.Token)
	resp, err = http.DefaultClient.Do(req)
	return
}

func jwtput(jwt auth.Token, url string, contenttype string, body *bytes.Buffer) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", jwt.Token)
	req.Header.Set("Content-Type", contenttype)
	resp, err = http.DefaultClient.Do(req)
	return
}

func jwtpostjson(token auth.Token, url string, body interface{}, result interface{}) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	resp, err := jwtpost(token, url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return
}

func jwtpost(jwt auth.Token, url string, contenttype string, body *bytes.Buffer) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", jwt.Token)
	req.Header.Set("Content-Type", contenttype)
	resp, err = http.DefaultClient.Do(req)
	return
}

func jwtget(jwt auth.Token, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", jwt.Token)
	resp, err = http.DefaultClient.Do(req)
	return
}

func createTestEnv(ctx context.Context, wg *sync.WaitGroup) (conf config.Config, err error) {
	conf, err = config.Load("../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		return conf, err
	}
	conf.MongoReplSet = false
	conf, err = NewDockerEnv(conf, ctx, wg)
	if err != nil {
		log.Println("ERROR: unable to create docker env", err)
		return conf, err
	}
	time.Sleep(1 * time.Second)
	err = Start(conf, ctx, wg)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		return conf, err
	}
	time.Sleep(10 * time.Second)
	return conf, err
}

func NewDockerEnv(startConfig config.Config, ctx context.Context, wg *sync.WaitGroup) (config config.Config, err error) {
	config = startConfig

	whPort, err := docker.GetFreePort()
	if err != nil {
		log.Println("unable to find free port", err)
		return config, err
	}
	config.ServerPort = strconv.Itoa(whPort)

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		return config, err
	}
	zkUrl := zkIp + ":2181"

	config.KafkaBootstrap, err = docker.Kafka(ctx, wg, zkUrl)
	if err != nil {
		return config, err
	}

	_, ip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		return config, err
	}
	config.MongoUrl = "mongodb://" + ip + ":27017"

	_, openseachIp, err := docker.OpenSearch(ctx, wg)
	if err != nil {
		return config, err
	}

	_, permIp, err := docker.PermissionSearch(ctx, wg, config.KafkaBootstrap, openseachIp)
	if err != nil {
		return config, err
	}
	config.PermissionsUrl = "http://" + permIp + ":8080"

	return config, nil
}
