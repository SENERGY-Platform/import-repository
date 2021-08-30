/*
 * Copyright 2018 InfAI (CC SES)
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

package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type JwtImpersonate struct {
	Token   string
	XUserId string
}

func (this *Token) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-UserId", this.Sub)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if err == nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		err = errors.New("access denied")
	}
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New("not found")
	}
	if resp.StatusCode == http.StatusInternalServerError {
		err = errors.New("internal server error")
	}
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println("DEBUG: ", url, resp.Status, resp.StatusCode, buf.String())
	}
	return
}

func (this *Token) Put(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-UserId", this.Sub)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if err == nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		err = errors.New("access denied")
	}
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New("not found")
	}
	if resp.StatusCode == http.StatusInternalServerError {
		err = errors.New("internal server error")
	}
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println("DEBUG: ", url, resp.Status, resp.StatusCode, buf.String())
	}
	return
}

func (this *Token) Delete(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("X-UserId", this.Sub)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if err == nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		err = errors.New("access denied")
	}
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New("not found")
	}
	if resp.StatusCode == http.StatusInternalServerError {
		err = errors.New("internal server error")
	}
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println("DEBUG: ", url, resp.Status, resp.StatusCode, buf.String())
	}
	return
}

func (this *Token) DeleteWithBody(url string, body interface{}) (resp *http.Response, err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	req, err := http.NewRequest("DELETE", url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("X-UserId", this.Sub)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if err == nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		err = errors.New("access denied")
	}
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New("not found")
	}
	if resp.StatusCode == http.StatusInternalServerError {
		err = errors.New("internal server error")
	}
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println("DEBUG: ", url, resp.Status, resp.StatusCode, buf.String())
	}
	return
}

func (this *Token) PostJSON(url string, body interface{}, result interface{}) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	resp, err := this.Post(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return
}

func (this *Token) PutJSON(url string, body interface{}, result interface{}) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	resp, err := this.Put(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return
}

func (this *Token) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("X-UserId", this.Sub)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if err == nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		err = errors.New("access denied")
	}
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New("not found")
	}
	if resp.StatusCode == http.StatusInternalServerError {
		err = errors.New("internal server error")
	}
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println("DEBUG: ", url, resp.Status, resp.StatusCode, buf.String())
	}
	return
}

func (this *Token) GetJSON(url string, result interface{}) (err error) {
	resp, err := this.Get(url)
	if err != nil {
		return err
	}
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(payload, result)
	return
}
