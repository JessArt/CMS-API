package main

import (
  "database/sql"
  "github.com/gocraft/dbr"
  "github.com/gin-gonic/gin"
  "strconv"
  "fmt"
)

func getTags(db *sql.DB) []gin.H {
  stmt, _ := db.Prepare("select name, id from tags;")

  defer stmt.Close()

  rows, _ := stmt.Query()

  defer rows.Close()

  var (
    tag string
    id string
    tags []gin.H
  )

  tags = make([]gin.H, 0)

  for rows.Next() {
    rows.Scan(&tag, &id)
    tags = append(tags, gin.H{
      "name": tag,
      "id": id,
    })
  }

  return tags
}

func getImageTags(sess *dbr.Session, imageID string) map[string]bool {
  result := make(map[string]bool)
  var tags []string
  sess.Select("tag_id").From("tags_images").Where("image_id = ?", imageID).Load(&tags)
  for _, tagID := range tags {
    result[tagID] = true
  }
  return result
}

func getTagIDByName(db *sql.DB, name string) string {
  stmt, err := db.Prepare("select id from tags where name = ?;")

  if err != nil {
    fmt.Println(err)
  }

  defer stmt.Close()
  rows, err := stmt.Query(name)

  defer rows.Close()

  var (
    id string
  )

  for rows.Next() {
    rows.Scan(&id)
  }

  return id
}

func createTags(db *sql.DB, tags []string) []string {
  newTags := make([]string, 0)
  for _, tag := range tags {
    // 1. check whether tags exists or not
    // if not, insert new row into tags table
    stmt, err := db.Prepare("select COUNT(*) from tags where id=?;")

    if err != nil {
      fmt.Println(err)
    }

    defer stmt.Close()

    rows, err := stmt.Query(tag)

    if err != nil {
      fmt.Println(err)
    }

    defer rows.Close()

    var (
      count int
    )

    for rows.Next() {
      rows.Scan(&count)

      if count == 0 {
        stmt, _ := db.Prepare("insert into tags (name) values (?);")

        defer stmt.Close()
        res, err := stmt.Exec(tag)

        if err != nil {
          fmt.Println(err)
        }

        id, _ := res.LastInsertId()
        newTags = append(newTags, strconv.FormatInt(id, 10))
      } else {
        newTags = append(newTags, tag)
      }
    }
  }

  return newTags
}

func updateImageTagsRelation(db *sql.DB, tags []string, imageID string) {
  deleteImageTags(db, imageID)
  createImageTagsRelation(db, tags, imageID)
}

func deleteImageTags(db *sql.DB, imageID string) {
  stmt, err := db.Prepare("delete from tags_images where image_id=?;")

  if err != nil {
    fmt.Println(err)
  }

  defer stmt.Close()

  stmt.Exec(imageID)
}

func createImageTagsRelation(db *sql.DB, tags []string, imageID string) {
  for _, tag := range tags {
    stmt, err := db.Prepare("insert into tags_images (tag_id, image_id) values (?, ?);")

    if err != nil {
      fmt.Println(err)
    }

    defer stmt.Close()

    // tagId := getTagIdByName(db, tag)

    stmt.Exec(tag, imageID)
  }
}
