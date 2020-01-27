package main

import (
  "net/http"
  //"errors"
  //"fmt"
  "github.com/gin-gonic/gin"
)

func robots_txt(c *gin.Context) {
  body := "User-agent: *\nDisallow: /\n"
  c.String(http.StatusOK, body)
}
