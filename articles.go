package main

import (
  "github.com/gin-gonic/gin"
  "github.com/gocraft/dbr"
  "net/http"
)

func getArticlesHandler(sess *dbr.Session, c *gin.Context) {
  type Article struct {
    ID string
    Title string
    Subtitle string
    Country string
    City string
  }
  var articles []Article
  sess.Select("id, title, subtitle").From("articles").Load(&articles)
  c.HTML(http.StatusOK, "articles", gin.H{
    "articles": articles,
  })
}

func getArticleHandler(c *gin.Context) {
  checkLogin(c)
  c.HTML(http.StatusOK, "article", gin.H{
    "title": "my title",
  })
}

func editArticleHandler(sess *dbr.Session, c *gin.Context, id string) {
  type Article struct {
    ID string
    Title string
    Subtitle string
    Cover string
    Country string
    City string
    Text string
    Keywords string
  }

  var article Article

  sess.Select("id, title, subtitle, cover, country, city, text, keywords").
    From("articles").Where("id = ?", id).Load(&article)

  c.HTML(http.StatusOK, "article", gin.H{
    "article": article,
  })
}

func saveArticle(sess *dbr.Session, c *gin.Context, folderFlag *string) {
  id := c.PostForm("id")

  title := c.PostForm("title")
  subtitle := c.PostForm("subtitle")
  cover := c.PostForm("cover")
  country := c.PostForm("country")
  city := c.PostForm("city")
  text, _ := fixText(c.PostForm("text"), folderFlag)
  keywords := c.PostForm("keywords")

  if id != "" {
    sess.Update("articles").
      Set("title", title).
      Set("subtitle", subtitle).
      Set("cover", cover).
      Set("country", country).
      Set("city", city).
      Set("text", text).
      Set("keywords", keywords).
      Where("id = ?", id).Exec()

    c.HTML(http.StatusOK, "success", gin.H{
      "message": "Your article was updated successfully!",
    })
  } else {
    sess.InsertInto("articles").
      Columns("title", "subtitle", "cover", "country", "city", "text", "keywords").
      Values(title, subtitle, cover, country, city, text, keywords).Exec()

    c.HTML(http.StatusOK, "success", gin.H{
      "message": "Your article was created successfully!",
    })
  }
}
