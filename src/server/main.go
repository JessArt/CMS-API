// +build linux,386

package main

import (
  "database/sql"
  "fmt"
  "flag"
  "github.com/anthonynsimon/bild/imgio"
  "github.com/anthonynsimon/bild/transform"
  "github.com/gin-gonic/gin"
  "github.com/satori/go.uuid"
  "image"
  "strconv"
  "net/http"

  _ "github.com/go-sql-driver/mysql"
  _ "image/jpeg"
  _ "image/png"
)

func savePath(folder string, filename string) string {
  return fmt.Sprintf("%s/%s", folder, filename)
}

func saveFile(folder string, filename string, imageFile image.Image) string {
  if err := imgio.Save(savePath(folder, filename), imageFile, imgio.JPEG); err != nil {
    panic(err)
  }

  return filename
}

func prepareURL(filename string) string {
  return "http://static.jess.gallery/" + filename
}

func main() {
  dbFlag := flag.String("db", "root@admin", "database credentials")
  nameFlag := flag.String("name", "test", "name of the database")
  folderFlag := flag.String("folder", "/static", "folder path to store images")

  flag.Parse()

  r := gin.Default()
  r.HTMLRender = createTemplates()

  dbAddress := fmt.Sprintf("%s@tcp(127.0.0.1:3306)/%s", *dbFlag, *nameFlag)

  fmt.Print(dbAddress)
  db, _ := sql.Open("mysql", dbAddress)

  defer db.Close()

  r.GET("/login", func(c *gin.Context) {
    c.HTML(http.StatusOK, "login", gin.H{
      "title": "My title",
    })
  })

  r.POST("/login", func(c *gin.Context) {
    loginHandler(db, c)
  })

  r.GET("/articles", func(c *gin.Context) {
    checkLogin(c)
    getArticlesHandler(db, c)
  })

  r.GET("/new/article", getArticleHandler)
  r.GET("/articles/:id", func(c *gin.Context) {
    checkLogin(c)
    id := c.Param("id")
    editArticleHandler(db, c, id)
  })

  r.POST("/article", func(c *gin.Context) {
    checkLogin(c)
    saveArticle(db, c)
  })

  r.GET("/new", func(c *gin.Context) {
    checkLogin(c)
    tags := getTags(db)
    fmt.Println(tags)
    c.HTML(http.StatusOK, "form", gin.H{
      "showImage": true,
      "title": "My title",
      "tags": getTags(db),
    })
  })

  r.GET("/", func(c *gin.Context) {
    checkLogin(c)

    c.HTML(http.StatusOK, "index", gin.H{
      "title": "Main page",
    })
  })

  r.GET("/images", func(c *gin.Context) {
    checkLogin(c)
    stmt, err := db.Prepare("select title, id, description, small_url from images;")

    if err != nil {
      panic(err)
    }

    defer stmt.Close()
    rows, err := stmt.Query()

    if err != nil {
      panic(err)
    }

    var (
      title string
      id string
      description string
      url string
      titles []gin.H
    )

    titles = make([]gin.H, 0)

    for rows.Next() {
      err := rows.Scan(&title, &id, &description, &url)
      if err != nil {
        fmt.Print(err.Error())
      }
      fmt.Println(title)
      titles = append(titles, gin.H{
        "title": title,
        "id": id,
        "description": description,
        "url": url,
      })
    }

    c.HTML(http.StatusOK, "images", gin.H{
      "images": titles,
    })
  })

  r.GET("/images/:id", func(c *gin.Context) {
    checkLogin(c)
    stmt, err := db.Prepare("select small_url, original_url, id, title, description, type, date, location from images where id =?;")

    if err != nil {
      fmt.Print(err.Error())
    }
    defer stmt.Close()
    id := c.Param("id")
    rows, err := stmt.Query(id)
    if err != nil {
      fmt.Print(err.Error())
    }

    tagsMap := getImageTags(db, id)
    fmt.Println(tagsMap)

    var (
      url string
      bigURL string
      imageId string
      title string
      description string
      image_type string
      date string
      location string
    )

    defer rows.Close()
    for rows.Next() {
      err := rows.Scan(&url, &bigURL, &imageId, &title, &description, &image_type, &date, &location)
      if err != nil {
        fmt.Print(err.Error())
      }

      c.HTML(http.StatusOK, "form", gin.H{
        "id": imageId,
        "url": url,
        "big_url": bigURL,
        "title": "My title",
        "isPhoto": image_type == "photo",
        "isArt": image_type == "art",
        "isOther": image_type == "other",
        "image_title": title,
        "description": description,
        "imageType": image_type,
        "date": date,
        "location": location,
        "tags": getTags(db),
        "currentTags": tagsMap,
      })
    }
  })

  r.POST("/image", func(c *gin.Context) {
    c.Request.ParseMultipartForm(5 * 1024 * 1024)
    tags := c.Request.Form["tags"]
    fmt.Println(tags)
    tags = createTags(db, tags)
    file, _, err := c.Request.FormFile("image")
    if err != nil {
      if id := c.PostForm("id"); id != "" {
        stmt, err := db.Prepare("UPDATE images SET title=?, type=?, description=?, date=?, location=? WHERE id=?;")
        if err != nil {
          fmt.Print(err.Error())
        }

        title := c.PostForm("title")
        typeField := c.PostForm("type")
        description := c.PostForm("description")
        date := c.PostForm("date")
        location := c.PostForm("location")

        _, err = stmt.Exec(title, typeField, description, date, location, id)

        if err != nil {
          fmt.Print(err.Error())
        }

        updateImageTagsRelation(db, tags, id)

        c.HTML(http.StatusOK, "success", gin.H{
          "message": fmt.Sprintf("Image %s was successfully updated", title),
        })
      } else {
        image_type := c.PostForm("type")
        currentTags := make(map[string]bool)
        for _, id := range tags {
          currentTags[id] = true
        }
        c.HTML(http.StatusOK, "form", gin.H{
          "title": "My title",
          "error": "Please, upload the image",
          "isPhoto": image_type == "photo",
          "isArt": image_type == "art",
          "isOther": image_type == "other",
          "image_title": c.PostForm("title"),
          "description": c.PostForm("description"),
          "image_type": image_type,
          "date": c.PostForm("date"),
          "location": c.PostForm("location"),
          "tags": getTags(db),
          "currentTags": currentTags,
        })
      }
    } else {
      filename := uuid.NewV4().String()
      imageFile, _, _ := image.Decode(file)
      b := imageFile.Bounds()
      fmt.Println(b.Max.X, b.Max.Y, b.Max.X/b.Max.Y)
      var imageRatio float64 = float64(b.Max.X) / float64(b.Max.Y)
      smallImage := transform.Resize(imageFile, 500, int(500/imageRatio), transform.Linear)
      largeImage := transform.Resize(imageFile, 1200, int(1200/imageRatio), transform.Linear)

      smallImageFilename := saveFile(*folderFlag, filename + "_500.jpg", smallImage)
      bigImageFilename := saveFile(*folderFlag, filename + "_1200.jpg", largeImage)
      originalImageFilename := saveFile(*folderFlag, filename + ".jpg", imageFile)

      stmt, err := db.Prepare(`
        INSERT INTO images
        (small_url, big_url, original_url, title, type, description, date, location, original_width, original_height)
        values(?,?,?,?,?,?,?,?,?,?);
      `)

      if err != nil {
        fmt.Print(err.Error())
      }

      title := c.PostForm("title")
      typeField := c.PostForm("type")
      description := c.PostForm("description")
      date := c.PostForm("date")
      location := c.PostForm("location")

      res, err := stmt.Exec(
        prepareURL(smallImageFilename),
        prepareURL(bigImageFilename),
        prepareURL(originalImageFilename),
        title,
        typeField,
        description,
        date,
        location,
        b.Max.X,
        b.Max.Y,
      )

      if err != nil {
        fmt.Print(err.Error())
      }

      imageID, err := res.LastInsertId()

      if err != nil {
        fmt.Println(err)
      }

      createImageTagsRelation(db, tags, strconv.FormatInt(imageID, 10))

      c.HTML(http.StatusOK, "success", gin.H{
        "message": "Image was successfully created",
      })
    }
  })

  r.OPTIONS("/v1/api/*action", preflightHandler)

  r.GET("/v1/api/images", func(c *gin.Context) {
    getImagesAPI(db, c)
  })

  r.Static("/assets", constructPath("static"))

  r.Run(":4002")
}
