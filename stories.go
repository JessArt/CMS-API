package main

import (
  "fmt"
  "encoding/json"
  "github.com/gin-gonic/gin"
  "github.com/gocraft/dbr"
  "net/http"
)

func getStoriesAPI(sess *dbr.Session, c *gin.Context) {
  setCORSHeaders(c)
  type Story struct {
    ID string
    Title string
    Subtitle string
    Cover string
  }

  var stories []Story

  sess.Select("*").From("stories").Load(&stories)

  c.JSON(http.StatusOK, stories)
}

func getStoryAPI(sess *dbr.Session, c *gin.Context, id string) {
  setCORSHeaders(c)
  type Image struct {
    ID string
    ImageID string
    Title string
    Description string
    Cover string
    Link string
  }

  type Story struct {
    ID string
    Title string
    Subtitle string
    Description string
    Cover string
    Keywords string `db:"meta_keywords"`
    MetaTitle string `db:"meta_title"`
    MetaDescription string `db:"meta_description"`
    Images []Image
  }

  var story Story

  sess.Select("*").From("stories").Where("id=?", id).Load(&story)

  var images []Image

  sess.Select("*").From("stories_images").Where("story_id=?", id).OrderBy("sort").Load(&images)

  story.Images = images

  c.JSON(http.StatusOK, story)
}

func listStories(sess *dbr.Session, c *gin.Context) {
  type Story struct {
    ID string
    Title string
    Cover string
  }

  var stories []Story

  sess.Select("*").From("stories").Load(&stories)

  c.HTML(http.StatusOK, "stories", gin.H{
    "title": "stories",
    "stories": stories,
  })
}

func getStoryHandler(c *gin.Context) {
  checkLogin(c)
  c.HTML(http.StatusOK, "story", gin.H{
    "title": "my title",
  })
}

func editStoryHandler(sess *dbr.Session, c *gin.Context, id string) {
  type Image struct {
    ID int
    ImageID string
    Title string
    Sort int
    Description string
    Cover string
    Link string
  }

  type Story struct {
    ID string
    Title string
    Subtitle string
    Description string
    Cover string
    Keywords string `db:"meta_keywords"`
    MetaTitle string `db:"meta_title"`
    MetaDescription string `db:"meta_description"`
    Images []Image
  }

  var story Story
  var images []Image

  sess.Select("*").From("stories").Where("id=?", id).Load(&story)
  sess.Select("*").From("stories_images").Where("story_id=?", id).OrderBy("sort").Load(&images)
  story.Images = images
  res, _ := json.Marshal(story)
  s := string(res)
  fmt.Println(s)
  c.HTML(http.StatusOK, "story", gin.H{
    "title": "my title",
    "json": s,
  })
}

func saveStory(sess *dbr.Session, c *gin.Context) {
  type Image struct {
    New bool `json:"new"`
    ID int `json:"id"`
    ImageID string `json:"imageId"`
    Title string `json:"title"`
    Sort int `json:"sort"`
    Description string `json:"description"`
    Cover string `json:"cover"`
    Link string `json:"link"`
    Remove bool `json:"remove"`
  }

  type Story struct {
    ID dbr.NullString `json:"id"`
    Title     string `form:"user" json:"title" binding:"required"`
    Subtitle string `json:"subtitle"`
    Description  string `form:"password" json:"description" binding:"required"`
    Cover string `json:"cover"`
    Keywords string `json:"keywords"`
    MetaTitle string `json:"metaTitle"`
    MetaDescription string `json:"metaDescription"`
    Images []Image `json:"images"`
  }

  var json Story
  c.BindJSON(&json)

  fmt.Println(json)

  if json.ID.Valid {
    sess.Update("stories").
      Set("title", json.Title).
      Set("subtitle", json.Subtitle).
      Set("cover", json.Cover).
      Set("description", json.Description).
      Set("meta_keywords", json.Keywords).
      Set("meta_title", json.MetaTitle).
      Set("meta_description", json.MetaDescription).
      Where("id = ?", json.ID.String).Exec()

    for _, image := range json.Images {
      if image.Remove == true {
        sess.DeleteFrom("stories_images").Where("id=?", image.ID).Exec()
      } else {
        if image.New == false {
          sess.Update("stories_images").
            Set("image_id", image.ImageID).
            Set("sort", image.Sort).
            Set("title", image.Title).
            Set("description", image.Description).
            Set("cover", image.Cover).
            Set("link", image.Link).
            Where("id=?", image.ID).Exec()
        } else {
          fmt.Println("saving new story image...")
          fmt.Println(image)
          _, err := sess.InsertInto("stories_images").
            Columns("image_id", "story_id", "sort", "title", "description", "cover", "link").
            Values(image.ImageID, json.ID.String, image.Sort, image.Title, image.Description, image.Cover, image.Link).Exec()

          fmt.Println(err)
        }
      }
    }

    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  } else {
    res, _ := sess.InsertInto("stories").
      Columns("title", "subtitle", "cover", "description", "meta_keywords", "meta_title", "meta_description").
      Values(json.Title, json.Subtitle, json.Cover, json.Description, json.Keywords, json.MetaTitle, json.MetaDescription).Exec()

    id, _ := res.LastInsertId()

    for _, image := range json.Images {
      sess.InsertInto("stories_images").
        Columns("image_id", "story_id", "sort", "title", "description", "cover", "link").
        Values(image.ImageID, id, image.Sort, image.Title, image.Description, image.Cover, image.Link).Exec()

      fmt.Println(id, image)
    }

    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  }
}
