package main

import (
  "fmt"
  "net/http"
  "database/sql"
  "github.com/gin-gonic/gin"
  "github.com/gocraft/dbr"
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
  imageType := c.DefaultQuery("type", "art")
  stmt, err := db.Prepare(`
    SELECT id, title, description, small_url, big_url, original_url, date, location, original_width, original_height
    FROM images
    WHERE type = ?
    ORDER BY id DESC
  `)
  defer stmt.Close()
  if err != nil {
    fmt.Print(err.Error())
  }

  rows, err := stmt.Query(imageType)

  var (
    id string
    title string
    description string
    smallURL string
    bigURL string
    originalURL string
    date string
    location string
    originalWidth int
    originalHeight int
  )

  defer rows.Close()
  content := make([]gin.H, 0)
  var counter int
  for rows.Next() {
    err := rows.Scan(&id, &title, &description, &smallURL, &bigURL, &originalURL, &date, &location, &originalWidth, &originalHeight)
    if err != nil {
      fmt.Print(err.Error())
    }

    content = append(content, gin.H{
      "title": title,
      "id": id,
      "type": imageType,
      "description": description,
      "small_url": smallURL,
      "big_url": bigURL,
      "original_url": originalURL,
      "date": date,
      "location": location,
      "originalWidth": originalWidth,
      "originalHeight": originalHeight,
    })

    counter = counter + 1
  }

  c.JSON(http.StatusOK, content)
}

func getImagesAPI2(sess *dbr.Session, c *gin.Context) {
  setCORSHeaders(c)
  imageType := c.DefaultQuery("type", "art")

  type Tag struct {
    Name string
  }

  type Image struct {
    Title string
    ID string
    Type string `db:"type"`
    Description dbr.NullString
    SmallURL string `db:"small_url"`
    BigURL string `db:"big_url"`
    OriginalURL string `db:"original_url"`
    Date dbr.NullString
    Location dbr.NullString
    Keywords dbr.NullString
    OriginalWidth dbr.NullInt64 `db:"original_width"`
    OriginalHeight dbr.NullInt64 `db:"original_height"`
    Tags []string
  }

  var images []Image
  imagesWithTags := make([]Image, 0)
  sess.Select("*").From("images").Where("type = ?", imageType).OrderDir("id", false).Load(&images)
  for _, image := range images {
    var tags []Tag
    sess.Select("*").From("tags_images").Join("tags", "tags_images.tag_id = tags.id").Where("tags_images.image_id = ?", image.ID).Load(&tags)
    plainTags := make([]string, 0)
    for _, tag := range tags {
      fmt.Println(tag.Name)
      plainTags = append(plainTags, tag.Name)
    }
    image.Tags = plainTags
    imagesWithTags = append(imagesWithTags, image)
  }

  c.JSON(http.StatusOK, imagesWithTags)
}

func getArticlesAPI(sess *dbr.Session, c *gin.Context) {
  setCORSHeaders(c)

  type Article struct {
    ID string
    Title dbr.NullString
    Subtitle dbr.NullString
    Cover dbr.NullString
    Country dbr.NullString
    City dbr.NullString
    Text dbr.NullString
    Keywords dbr.NullString
  }

  var articles []Article
  sess.Select("*").From("articles").OrderDir("id", false).Load(&articles)

  c.JSON(http.StatusOK, articles)
}
