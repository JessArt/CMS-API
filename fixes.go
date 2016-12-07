package main

import (
  "fmt"
  "strings"
  "net/http"
  "regexp"
  "image"
  "github.com/anthonynsimon/bild/transform"
  "github.com/satori/go.uuid"
  "github.com/gocraft/dbr"

  _ "image/jpeg"
  _ "image/png"
)

func fixURL(url string) string {
  startsHTTP := strings.HasPrefix(url, "http:")

  if startsHTTP == true {
    return url[5:]
  }

  return url
}

func fixLinks(sess *dbr.Session) {
  type Image struct {
    ID string
    SmallURL string `db:"small_url"`
    BigURL string `db:"big_url"`
    OriginalURL string `db:"original_url"`
  }

  var images []Image
  sess.Select("id, small_url, big_url, original_url").From("images").Load(&images)

  for _, image := range images {
    fmt.Println(fixURL(image.SmallURL))
    sess.Update("images").
      Set("small_url", fixURL(image.SmallURL)).
      Set("big_url", fixURL(image.BigURL)).
      Set("original_url", fixURL(image.OriginalURL)).
      Where("id = ?", image.ID).Exec()
  }
}

func fixTags(sess *dbr.Session) {
  type Tag struct {
    Name string
    Number int
  }
  var tags []Tag

  sess.SelectBySql("SELECT `name`, COUNT(*) as `number` from `tags` GROUP BY `name` HAVING `number` > 1").Load(&tags)

  fmt.Println("==============")
  fmt.Println(len(tags))
  fmt.Println("==============")

  for _, tag := range tags {
    var similarTags []string
    sess.Select("id").From("tags").Where("name = ?", tag.Name).Load(&similarTags)

    first := similarTags[0]

    fmt.Println(first, similarTags[1:], tag.Name)

    sess.Update("tags_images").Set("tag_id", first).Where("tag_id IN ?", similarTags[1:]).Exec()

    sess.DeleteFrom("tags").Where("id IN ?", similarTags[1:]).Exec()
  }
}

func fixText(text string, folderFlag *string) (string, int) {
  // find <img src="$EXT" />, where $EXT leads to other articles
  // yes, I know the zen – https://blog.codinghorror.com/parsing-html-the-cthulhu-way/
  r, _ := regexp.Compile("<img(.+?)src=\"(.+?)\"")
  res := r.FindAllStringSubmatch(text, -1)

  type Change struct {
    Old string
    New string
  }

  var changes = make([]Change, 0)

  for _, value := range res {
    filename := uuid.NewV4().String()
    url := value[2]

    hasHTTP := strings.HasPrefix(url, "http://static.jess.gallery")
    hasHTTPS := strings.HasPrefix(url, "https://static.jess.gallery")
    alreadyFixed := strings.HasPrefix(url, "//static.jess.gallery")

    if hasHTTP {
      newURL := strings.Replace(url, "http://static.jess.gallery", "//static.jess.gallery", 1)
      changes = append(changes, Change{ Old: url, New: newURL })
    } else if hasHTTPS {
      newURL := strings.Replace(url, "https://static.jess.gallery", "//static.jess.gallery", 1)
      changes = append(changes, Change{ Old: url, New: newURL })
    } else if alreadyFixed != true {
      response, _ := http.Get(url)

      defer response.Body.Close()
      imageFile, _, _ := image.Decode(response.Body)

      b := imageFile.Bounds()
      fmt.Println(b.Max.X, b.Max.Y, b.Max.X/b.Max.Y)
      var imageRatio = float64(b.Max.X) / float64(b.Max.Y)
      // compress them to 1200px width and save to the static folder
      largeImage := transform.Resize(imageFile, 1200, int(1200/imageRatio), transform.Linear)
      finalLastName := saveFile(*folderFlag, filename + "_1200.jpg", largeImage)
      fmt.Println(finalLastName, largeImage.Bounds())

      newURL := prepareURL(finalLastName)

      changes = append(changes, Change{ Old: url, New: newURL })
    }
  }

  articleText := text
  // change link to the internal one
  if len(changes) > 0 {
    for _, change := range changes {
      articleText = strings.Replace(articleText, change.Old, change.New, 1)
    }
  }

  return articleText, len(changes)
}

func fixExternalImages(sess *dbr.Session, folderFlag *string) {
  type Article struct {
    ID string
    Text string
  }
  // load all articles
  var articles []Article
  sess.Select("id, text").From("articles").Load(&articles)
  // iterate through them and get their text
  for _, article := range articles {
    newText, numberOfChanges := fixText(article.Text, folderFlag)

    if numberOfChanges > 0 {
      sess.Update("articles").Set("text", newText).Where("id = ?", article.ID).Exec()
    }
  }
}
