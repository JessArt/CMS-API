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
