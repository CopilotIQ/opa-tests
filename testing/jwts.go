// Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
//
// Created by M. Massenzio, 2022-06-22

package testing

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/massenz/slf4go/logging"
)

var (
	SecretKey = []byte("doe5n7matter")
)

func NewToken(body *JwtBody) string {
	logging.RootLog.Debug("Creating JWT with body: %v", body)
	claims := jwt.MapClaims{
		"sub":     body.Subject,
		"user_id": body.UserId,
		"roles":   body.Roles,
		"iss":     body.Issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString(SecretKey)
	return ss
}
