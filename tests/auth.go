/*
 * Copyright 2024 InfAI (CC SES)
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
	"strings"
	"time"

	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	gojwt "github.com/golang-jwt/jwt"
)

type RealmAccess struct {
	Roles []string `json:"roles"`
}

type KeycloakClaims struct {
	RealmAccess RealmAccess `json:"realm_access"`
	gojwt.StandardClaims
}

func createToken(issuer string, userId string) (token jwt.Token, err error) {
	return createTokenWithRoles(issuer, userId, []string{})
}

func createTokenWithRoles(issuer string, userId string, roles []string) (token jwt.Token, err error) {
	realmAccess := RealmAccess{Roles: roles}
	claims := KeycloakClaims{
		realmAccess,
		gojwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			Issuer:    issuer,
			Subject:   userId,
		},
	}

	jwtoken := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	unsignedTokenString, err := jwtoken.SigningString()
	if err != nil {
		return token, err
	}
	tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
	token.Token = "Bearer " + tokenString
	token.Sub = userId
	token.RealmAccess = map[string][]string{"roles": roles}
	return token, nil
}
