package main

import (
  "fmt"
  "database/sql"
  "github.com/gin-gonic/gin"
  "net/http"
)

func getArticlesHandler(db *sql.DB, c *gin.Context) {
  stmt, err := db.Prepare("select id, title, subtitle from articles")

  if err != nil {
    fmt.Println(err)
  }

  defer stmt.Close()

  rows, err := stmt.Query()

  if err != nil {
    fmt.Println(err)
  }

  defer rows.Close()

  var (
    id, title, subtitle string
  )

  articles := make([]gin.H, 0)

  for rows.Next() {
    rows.Scan(&id, &title, &subtitle)

    articles = append(articles, gin.H{
      "id": id,
      "title": title,
      "subtitle": subtitle,
    })
  }

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

func editArticleHandler(db *sql.DB, c *gin.Context, id string) {
  stmt, err := db.Prepare("select title, subtitle, cover, country, city, text from articles where id=?;")

  if err != nil {
    fmt.Println(err)
  }

  defer stmt.Close()
  rows, err := stmt.Query(id)

  if err != nil {
    fmt.Println(err)
  }

  var (
    title string
    subtitle string
    cover string
    country string
    city string
    text string
  )

  defer rows.Close()
  for rows.Next() {
    rows.Scan(&title, &subtitle, &cover, &country, &city, &text)

    c.HTML(http.StatusOK, "article", gin.H{
      "id": id,
      "title": title,
      "subtitle": subtitle,
      "cover": cover,
      "country": country,
      "city": city,
      "text": text,
    })
  }
}

func saveArticle(db *sql.DB, c *gin.Context) {
  id := c.PostForm("id")

  title := c.PostForm("title")
  subtitle := c.PostForm("subtitle")
  cover := c.PostForm("cover")
  country := c.PostForm("country")
  city := c.PostForm("city")
  text := c.PostForm("text")

  if id != "" {
    // need to update
    stmt, err := db.Prepare("UPDATE articles SET title=?, subtitle=?, cover=?, country=?, city=?, text=? where id=?;")

    if err != nil {
      fmt.Println(err)
    }

    defer stmt.Close()

    _, err = stmt.Exec(title, subtitle, cover, country, city, text, id)

    if err != nil {
      fmt.Println(err)
    } else {
      c.HTML(http.StatusOK, "success", gin.H{
        "message": "Your article was updated successfully!",
      })
    }
  } else {
    // need to save new entry
    stmt, err := db.Prepare("insert into articles (title, subtitle, cover, country, city, text) values (?, ?, ?, ?, ?, ?);")

    if err != nil {
      fmt.Println(err)
    }

    defer stmt.Close()

    _, err = stmt.Exec(title, subtitle, cover, country, city, text)

    if err != nil {
      fmt.Println(err)
    } else {
      c.HTML(http.StatusOK, "success", gin.H{
        "message": "Your article was created successfully!",
      })
    }
  }
}
