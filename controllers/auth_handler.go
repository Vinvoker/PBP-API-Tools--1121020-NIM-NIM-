package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func generateToken(c *gin.Context, id int, name string, userType int) {
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")

	expiryTime := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &CustomClaims{
		ID:       id,
		Name:     name,
		UserType: userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiryTime.Unix(),
		},
	})

	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  expiryTime,
		Secure:   false,
		HttpOnly: true,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookie, err := c.Cookie("Authorization"); err == nil {
			if cookie == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			jwtSecretKey := os.Getenv("JWT_SECRET_KEY")

			parsedToken, err := jwt.ParseWithClaims(cookie, &CustomClaims{}, func(accessToken *jwt.Token) (interface{}, error) {
				return []byte(jwtSecretKey), nil
			})

			if err != nil || !parsedToken.Valid {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			claims, ok := parsedToken.Claims.(*CustomClaims)
			if !ok || claims.ID == 0 {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Next()
		}
	}
}
