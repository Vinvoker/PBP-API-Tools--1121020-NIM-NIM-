package controllers

import "github.com/dgrijalva/jwt-go"

type User struct {
	UserId       int    `json:"id"`
	UserName     string `json:"name"`
	UserEmail    string `json:"email"`
	UserCountry  string `json:"country"`
	UserType     string `json:"type"`
	userPassword string
}

type CustomClaims struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	UserType int    `json:"userType"`
	jwt.StandardClaims
}
