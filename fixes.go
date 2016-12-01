package main

import (
  "fmt"
  "strings"
  "github.com/gocraft/dbr"
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
