package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})
var ctx = context.Background()

func ActivateCRON(c *gin.Context) {
	minutes := c.Param("minutes")

	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatal(err)
	}
	s := gocron.NewScheduler(location)
	s.Every(minutes).Seconds().Do(func() {
		SendEmail(c)
	})
	s.StartAsync()
}

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
	var wg sync.WaitGroup
	wg.Add(len(emails))
	d := gomail.NewDialer("smtp.gmail.com", 587, env_email, env_password)

	for i, v := range names {
		go func(i int, v string) {
			defer wg.Done()

			m := gomail.NewMessage()
			m.SetHeader("From", "if-21020@students.ithb.ac.id")
			m.SetHeader("To", emails[i])
			m.SetHeader("Subject", "Hello!")
			m.SetBody("text/html", "Hello "+v)

			if err := d.DialAndSend(m); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}(i, v)
	}

	wg.Wait()
	c.JSON(http.StatusOK, gin.H{"message": "All emails send successfully"})
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
