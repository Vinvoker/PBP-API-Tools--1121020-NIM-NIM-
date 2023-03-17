package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Secure:   false,
		HttpOnly: true,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func Login(c *gin.Context) {
	db := connect()
	defer db.Close()

	input_email := c.PostForm("email")
	input_password := c.PostForm("password")

	stmt, err := db.Prepare("SELECT userpassword, userid, username, usertype FROM users WHERE useremail = ?")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer stmt.Close()

	var userid int
	var username string
	var usertype int
	var password string
	err = stmt.QueryRow(input_email).Scan(&password, &userid, &username, &usertype)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if password == input_password {
		generateToken(c, userid, username, usertype)
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

func GetUsers(c *gin.Context) {
	var users []User

	db := connect()
	defer db.Close()

	rows, err := db.Query("SELECT userid, username, useremail, usercountry, usertype FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserId, &user.UserName, &user.UserEmail, &user.UserCountry, &user.UserType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindWith(&user, binding.Form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := connect()
	defer db.Close()

	result, err := db.Exec("INSERT INTO users (username, useremail, usercountry, usertype, userpassword) VALUES (?, ?, ?, ?, ?)", user.UserName, user.UserEmail, user.UserCountry, user.UserType, user.userPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.UserId = int(id)
	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user User
	if err := c.ShouldBindWith(&user, binding.Form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.UserId = userId

	db := connect()
	defer db.Close()

	_, err = db.Exec("UPDATE users SET username=?, useremail=?, usercountry=?, usertype=?, userpassword=? WHERE userid=?", user.UserName, user.UserEmail, user.UserCountry, user.UserType, user.userPassword, user.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	db := connect()
	defer db.Close()

	_, err = db.Exec("DELETE FROM users WHERE userid=?", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
