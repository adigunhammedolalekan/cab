package models

import "github.com/dgrijalva/jwt-go"

type Token struct {
	jwt.StandardClaims
	UserId uint
}
