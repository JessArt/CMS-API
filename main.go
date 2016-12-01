package main

import (
  "database/sql"
  "fmt"
  "flag"
  "github.com/gocraft/dbr"
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
  return "//static.jess.gallery/" + filename
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
  conn, _ := dbr.Open("mysql", dbAddress, nil)

  sess := conn.NewSession(nil)
  fixLinks(sess)
  fixTags(sess)

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
    sess := conn.NewSession(nil)
    checkLogin(c)
    getArticlesHandler(sess, c)
  })

  r.GET("/new/article", getArticleHandler)
  r.GET("/articles/:id", func(c *gin.Context) {
    sess := conn.NewSession(nil)
    checkLogin(c)
    id := c.Param("id")
    editArticleHandler(sess, c, id)
  })

  r.POST("/article", func(c *gin.Context) {
    sess := conn.NewSession(nil)
    checkLogin(c)
    saveArticle(sess, c)
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
    imageType := c.DefaultQuery("type", "")
    sess := conn.NewSession(nil)

    type Image struct {
      ID string
      Title string
      Description string
      URL string `db:"small_url"`
    }

    var images []Image
    builder := sess.Select("id", "title", "description", "small_url").From("images")
    if imageType != "" {
      builder.Where("type = ?", imageType)
    }

    builder.Load(&images)
    c.HTML(http.StatusOK, "images", gin.H{
      "type": imageType,
      "images": images,
    })
  })

  r.GET("/images/:id", func(c *gin.Context) {
    checkLogin(c)
    sess := conn.NewSession(nil)

    id := c.Param("id")
    type Image struct {
      ID string
      Title string
      Description string
      SmallURL string `db:"small_url"`
      OriginalURL string `db:"original_url"`
      Type string
      Date string
      Location string
      Keywords string
    }

    var image Image

    sess.Select(
      "id",
      "title",
      "description",
      "small_url",
      "original_url",
      "type", "date", "location", "keywords",
    ).From("images").Where("id = ?", id).Load(&image)

    tagsMap := getImageTags(sess, id)
    fmt.Println(tagsMap)

    c.HTML(http.StatusOK, "form", gin.H{
      "Image": image,
      "isPhoto": image.Type == "photo",
      "isArt": image.Type == "art",
      "isCraft": image.Type == "craft",
      "isPostcard": image.Type == "postcard",
      "isOther": image.Type == "other",
      "tags": getTags(db),
      "currentTags": tagsMap,
    })
  })

  r.POST("/image", func(c *gin.Context) {
    c.Request.ParseMultipartForm(5 * 1024 * 1024)
    tags := c.Request.Form["tags"]
    fmt.Println(tags)
    tags = createTags(db, tags)
    file, _, err := c.Request.FormFile("image")
    if err != nil {
      if id := c.PostForm("id"); id != "" {
        stmt, err := db.Prepare("UPDATE images SET title=?, type=?, description=?, date=?, location=?, keywords=? WHERE id=?;")
        if err != nil {
          fmt.Print(err.Error())
        }

        title := c.PostForm("title")
        typeField := c.PostForm("type")
        description := c.PostForm("description")
        date := c.PostForm("date")
        location := c.PostForm("location")
        keywords := c.PostForm("keywords")

        _, err = stmt.Exec(title, typeField, description, date, location, keywords, id)

        if err != nil {
          fmt.Print(err.Error())
        }

        updateImageTagsRelation(db, tags, id)

        c.HTML(http.StatusOK, "success", gin.H{
          "message": fmt.Sprintf("Image %s was successfully updated", title),
        })
      } else {
        imageType := c.PostForm("type")
        currentTags := make(map[string]bool)
        for _, id := range tags {
          currentTags[id] = true
        }
        c.HTML(http.StatusOK, "form", gin.H{
          "title": "My title",
          "error": "Please, upload the image",
          "Image": gin.H{
            "Title": c.PostForm("title"),
            "Description": c.PostForm("description"),
            "Type": imageType,
            "Date": c.PostForm("date"),
            "Location": c.PostForm("location"),
            "Keywords": c.PostForm("keywords"),
          },
          "isPhoto": imageType == "photo",
          "isArt": imageType == "art",
          "isCraft": imageType == "craft",
          "isPostcard": imageType == "postcard",
          "isOther": imageType == "other",
          "tags": getTags(db),
          "currentTags": currentTags,
        })
      }
    } else {
      filename := uuid.NewV4().String()
      imageFile, _, _ := image.Decode(file)
      b := imageFile.Bounds()
      fmt.Println(b.Max.X, b.Max.Y, b.Max.X/b.Max.Y)
      var imageRatio = float64(b.Max.X) / float64(b.Max.Y)
      smallImage := transform.Resize(imageFile, 500, int(500/imageRatio), transform.Linear)
      largeImage := transform.Resize(imageFile, 1200, int(1200/imageRatio), transform.Linear)

      smallImageFilename := saveFile(*folderFlag, filename + "_500.jpg", smallImage)
      bigImageFilename := saveFile(*folderFlag, filename + "_1200.jpg", largeImage)
      originalImageFilename := saveFile(*folderFlag, filename + ".jpg", imageFile)

      stmt, err := db.Prepare(`
        INSERT INTO images
        (small_url, big_url, original_url, title, type, description, date, location, original_width, original_height, keywords)
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
      keywords := c.PostForm("keywords")

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
        keywords,
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

  r.POST("/v1/api/feedback", handleFeedback)

  r.GET("/v1/api/images", func(c *gin.Context) {
    getImagesAPI(db, c)
  })

  r.GET("/v1/api/articles", func(c *gin.Context) {
    sess := conn.NewSession(nil)
    getArticlesAPI(sess, c)
  })

  r.Static("/assets", constructPath("static"))

  r.Run(":4002")
}
