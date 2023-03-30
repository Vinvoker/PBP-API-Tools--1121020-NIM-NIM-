package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
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
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatal(err)
	}
	s := gocron.NewScheduler(location)
	// s.Every(1).Minutes().Do(func() {
	// 	SendEmail(c)
	// })
	s.Every(1).Day().At("09.00").Do(func() {
		SendEmail(c)
	})
	s.StartAsync()
}

func GetOwners(c *gin.Context) []Receiver {

	var receiver Receiver
	var receivers []Receiver

	db := connect()
	defer db.Close()

	rows, err := db.Query("SELECT ownerName, ownerEmail from dataowners")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "Something has gone wrong with dataowner the query"})
		return receivers
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&receiver.OwnerName, &receiver.OwnerEmail); err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": "Products not found"})
			return receivers
		} else {
			receivers = append(receivers, receiver)
		}
	}
	return receivers
}

func RekapOrder(c *gin.Context) []Message {
	db := connect()
	defer db.Close()

	var messages []Message
	var message Message

	rows, err := db.Query("SELECT p.productName, SUM(od.quantity), p.price FROM OrderDetails od" +
		" JOIN `Order` o ON o.orderId = od.orderId JOIN Product p ON p.productId=od.productId WHERE" +
		" o.transactionTime >= DATE(NOW() - INTERVAL 1 DAY) AND o.transactionTime < DATE(NOW()) GROUP BY p.productId")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "Something has gone wrong with the rekap order query"})
		return messages
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&message.ProductName, &message.Quantity, &message.Price); err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"error": "product not found"})
		} else {
			messages = append(messages, message)
		}
	}

	return messages

}

func GetValueFromRedis(productName string) string {
	// cek key-value nya masih ada atau sudah expire
	// kalau expirenya dalam waktu 3 menit atau kurang, refresh cache
	if rdb.TTL(ctx, productName).Val() < 180 {
		CacheProdukGambar()
	}
	res, err := rdb.Get(ctx, productName).Result()
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func SendEmail(c *gin.Context) {
	env_email := os.Getenv("EMAIL")
	env_password := os.Getenv("PASSWORD")

	receivers := GetOwners(c)
	messages := RekapOrder(c)

	var wg sync.WaitGroup
	wg.Add(len(receivers))
	d := gomail.NewDialer("smtp.gmail.com", 587, env_email, env_password)

	// bikin body string disini
	grandTotal := 0
	var strData string
	for _, message := range messages {
		totalProduct := message.Quantity * message.Price
		grandTotal += totalProduct
		productPicture := GetValueFromRedis(message.ProductName)

		strData += "<br><br>Product Name: " + message.ProductName
		strData += "<br>Product Price: Rp" + strconv.Itoa(message.Price)
		strData += "<br>Quantity of items bought: " + strconv.Itoa(message.Quantity)
		strData += "<br>Total from Products: Rp" + strconv.Itoa(totalProduct)
		strData += "<br> <img src='" + productPicture + "' alt='" + message.ProductName + "' width='200' height='300'/>"
	}

	stringBody := "<br>Rekap Harian Tanggal " + time.Now().AddDate(0, 0, -1).Format("02-01-2006")
	stringBody += strData
	stringBody += "<br><br><b>Grand Total : Rp" + strconv.Itoa(grandTotal) + "</b>"

	for i, receiver := range receivers {
		go func(i int, receiver Receiver) {
			defer wg.Done()
			m := gomail.NewMessage()
			m.SetHeader("From", "if-21020@students.ithb.ac.id")
			m.SetHeader("To", receiver.OwnerEmail)
			m.SetHeader("Subject", "Rekap Penjualan Harian Fore Cafe")
			m.SetBody("text/html", "Selamat pagi, "+receiver.OwnerName+stringBody)

			if err := d.DialAndSend(m); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}(i, receiver)

	}
	wg.Wait()
	c.JSON(http.StatusOK, gin.H{"message": "All emails successfully sent"})
}

func CacheProdukGambar() {
	db := connect()
	defer db.Close()

	rows, err := db.Query("SELECT productname, productpicture FROM product")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var picture string
	var name string
	for rows.Next() {
		if err := rows.Scan(&name, &picture); err != nil {
			panic(err)
		}
		err = rdb.Set(ctx, name, picture, 24*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}
}
