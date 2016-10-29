package main

import (
  "fmt"
  "net/http"
  "database/sql"
  "github.com/gin-gonic/gin"
)

func setCORSHeaders(c *gin.Context) {
  c.Header("Access-Control-Allow-Origin", "*")
  c.Header("Access-Control-Allow-Credentials", "true")
  c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
  c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
}

func preflightHandler (c *gin.Context) {
  setCORSHeaders(c)
  c.JSON(http.StatusOK, struct{}{})
}

func getImagesAPI(db *sql.DB, c *gin.Context) {
  setCORSHeaders(c)
  // decoder := json.NewDecoder(c.Request.Body)
  // type params struct {
  //   Type string
  // }
  // var t params
  // err := decoder.Decode(&t)
  // if err != nil {
  //     panic(err)
  // }

  imageType := c.DefaultQuery("type", "art")

  stmt, err := db.Prepare("select title, description, small_url, big_url, original_url, date, location from images where type = ?")
  defer stmt.Close()
  if err != nil {
    fmt.Print(err.Error())
  }

  rows, err := stmt.Query(imageType)

  var (
    title string
    description string
    smallURL string
    bigURL string
    originalURL string
    date string
    location string
  )

  defer rows.Close()
  content := make([]gin.H, 0)
  var counter int
  for rows.Next() {
    err := rows.Scan(&title, &description, &smallURL, &bigURL, &originalURL, &date, &location)
    if err != nil {
      fmt.Print(err.Error())
    }

    content = append(content, gin.H{
      "title": title,
      "description": description,
      "small_url": smallURL,
      "big_url": bigURL,
      "original_url": originalURL,
      "date": date,
      "location": location,
    })

    counter = counter + 1
  }

  c.JSON(http.StatusOK, content)
}
