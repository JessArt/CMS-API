package main

import (
  "github.com/gin-gonic/contrib/renders/multitemplate"
  "fmt"
  "os"
  "path/filepath"
)

func constructPath(path string) string {
  // from here – http://stackoverflow.com/questions/18537257/golang-how-to-get-the-directory-of-the-currently-running-file
  // because we can process from anywhere
  dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
  if err != nil {
    panic(err)
  }
  fmt.Println(dir)
  fmt.Println(dir)
  return fmt.Sprintf("%s/%s", dir, path)
}

func addTemplate(templates multitemplate.Render, name string, template string) {
  templatePath := fmt.Sprintf("templates/pages/%s.tmpl", template)
  templates.AddFromFiles(name,
    constructPath("templates/layouts/default.tmpl"),
    constructPath(templatePath))
}

func createTemplates() multitemplate.Render {
  templates := multitemplate.New()
  addTemplate(templates, "form", "template")
  addTemplate(templates, "index", "index")
  addTemplate(templates, "images", "images")
  addTemplate(templates, "success", "success")
  addTemplate(templates, "login", "login")
  addTemplate(templates, "article", "article")
  addTemplate(templates, "articles", "articles")
  addTemplate(templates, "subscribers", "subscribers")
  addTemplate(templates, "stories", "stories")
  addTemplate(templates, "story", "story")

  return templates
}
