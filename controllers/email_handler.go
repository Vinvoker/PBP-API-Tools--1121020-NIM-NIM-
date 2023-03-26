package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})
var ctx = context.Background()

func SendEmail(c *gin.Context) {
	country := c.Param("country")
	UpdateEmailList(c, country)

	key := "emailFrom" + country
	res, err := rdb.Get(ctx, key).Result()
	if err != nil {
		log.Fatal(err)
	}

	emails := []string{}
	err = json.Unmarshal([]byte(res), &emails)
	if err != nil {
		log.Fatal(err)
	}

	key = "nameFrom" + country
	res, err = rdb.Get(ctx, key).Result()
	if err != nil {
		log.Fatal(err)
	}

	names := []string{}
	err = json.Unmarshal([]byte(res), &names)
	if err != nil {
		log.Fatal(err)
	}

	env_email := os.Getenv("EMAIL")
	env_password := os.Getenv("PASSWORD")
	for i, v := range names {
		m := gomail.NewMessage()
		m.SetHeader("From", "if-21020@students.ithb.ac.id")
		m.SetHeader("To", emails[i])
		m.SetHeader("Subject", "Hello!")
		m.SetBody("text/html", "Hello "+v)

		d := gomail.NewDialer("smtp.gmail.com", 587, env_email, env_password)

		if err := d.DialAndSend(m); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
}

func UpdateEmailList(c *gin.Context, country string) {
	db := connect()
	defer db.Close()

	stmt, err := db.Prepare("SELECT username, useremail FROM users WHERE usercountry = ?")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	arrName := []string{}
	var name string
	arrEmail := []string{}
	var email string
	for rows.Next() {
		if err := rows.Scan(&name, &email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		arrName = append(arrName, name)
		arrEmail = append(arrEmail, email)
	}
	key := "emailFrom" + country
	data, err := json.Marshal(arrEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = rdb.Set(ctx, key, data, 0).Err()
	if err != nil {
		panic(err)
	}
	key = "nameFrom" + country
	data, err = json.Marshal(arrName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = rdb.Set(ctx, key, data, 0).Err()
	if err != nil {
		panic(err)
	}
}
