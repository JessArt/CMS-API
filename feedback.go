package main

import (
  "fmt"
  "net/http"
  "gopkg.in/gomail.v2"
  "github.com/gin-gonic/gin"
)

type Feedback struct {
    Name     string `form:"name" json:"name" binding:"required"`
    Email    string `form:"email" json:"email"`
    Message  string `form:"message" json:"message" binding:"required"`
}

func handleFeedback(c *gin.Context) {
  setCORSHeaders(c)
  var feedback Feedback

  if c.BindJSON(&feedback) == nil {
    m := gomail.NewMessage()
    m.SetHeader("From", "feedback.jess.gallery@gmail.com")
    m.SetHeader("To", "jess.zaikova@gmail.com")
    m.SetAddressHeader("Cc", "seva.zaikov@gmail.com", "Seva")
    m.SetHeader("Subject", "Response from a webpage!!!!")
    body := "From: " + feedback.Name + ";\n\n"
    body = body + "Email: " + feedback.Email + ";\n\n"
    body = body + "Message:\n" + feedback.Message
    m.SetBody("text/plain", body)

    d := gomail.NewDialer("smtp.gmail.com", 587, "feedback.jess.gallery@gmail.com", "o3rly8q8xc")

    if err := d.DialAndSend(m); err != nil {
      panic(err)
    }

    c.JSON(http.StatusOK, gin.H{
      "title": "!!!!",
    })
  } else {
    fmt.Println("Something is wrong...")
    fmt.Println(feedback)
  }
}
