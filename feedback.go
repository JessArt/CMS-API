package main

import (
  "fmt"
  "net/http"
  "gopkg.in/gomail.v2"
  "github.com/gin-gonic/gin"
  "github.com/gocraft/dbr"
)

func handleFeedback(c *gin.Context) {
  setCORSHeaders(c)

  type Feedback struct {
      Name     string `form:"name" json:"name" binding:"required"`
      Email    string `form:"email" json:"email"`
      Message  string `form:"message" json:"message" binding:"required"`
  }

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

func listSubscribers(sess *dbr.Session, c *gin.Context) {
  type Subscriber struct {
    Email dbr.NullString
  }

  var subscribers []Subscriber

  sess.Select("email").From("subscriptions").Load(&subscribers)

  vsf := make([]string, 0)
  for _, v := range subscribers {
    if v.Email.Valid && v.Email.String != "" {
      vsf = append(vsf, v.Email.String)
    }
  }

  c.HTML(http.StatusOK, "subscribers", gin.H{
    "title": "Subscribers",
    "subscribers": vsf,
  })
}

func addSubscription(sess *dbr.Session, c *gin.Context, email string) {
  setCORSHeaders(c)
  sess.InsertInto("subscriptions").Columns("email").Values(email).Exec()
  c.JSON(http.StatusCreated, nil)
}

func removeSubscription(sess *dbr.Session, c *gin.Context, email string) {
  setCORSHeaders(c)
  sess.DeleteFrom("subscriptions").Where("email = ?", email).Exec()

  c.JSON(http.StatusNoContent, nil)
}
