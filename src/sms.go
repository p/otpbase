package main

import (
  //"errors"
  "fmt"
  "github.com/gin-gonic/gin"
  //"os"
  //"strconv"
  "sync"
  //"io"
  //"io/ioutil"
  //"log"
  //"regexp"
  //"strings"
  "net/http"
  "time"
)

const CODE_REGEXP = "(\\d{6,})"

var mutex *sync.Mutex

type entry struct {
  text        string
  received_at int64
}

var entries []entry

func twiml(forward_number string, from_number string, text string) string {
  twiml_template := `
<?xml version='1.0' encoding='UTF-8'?>
<Response>
    <Message to='%s'>[OTPBASE:%s] %s</Message>
</Response>
`
  return fmt.Sprintf(twiml_template, forward_number, from_number, text)
}

var forward_number string

func expire_sms() {
  for range ticker.C {
    mutex.Lock()
    for len(entries) > 0 {
      entry := entries[len(entries)-1]
      if time.Now().UnixNano()-entry.received_at > 60e9 {
        entries = entries[:len(entries)-1]
      } else {
        break
      }
    }
    mutex.Unlock()
  }
}

func receive_sms(c *gin.Context) {
  text := c.PostForm("Body")
  from_number := c.PostForm("From")

  if len(text) == 0 {
    //c.AbortWithError(400, errors.New("Empty body is not allowed"))
    c.String(400, "Empty body is not allowed")
    return
  }

  mutex.Lock()
  entries = append([]entry{entry{text, time.Now().UnixNano()}}, entries...)
  if len(entries) > 5 {
    entries = entries[:5]
  }
  mutex.Unlock()

  if forward_number != "" {
    resp := twiml(forward_number, from_number, text)
    c.Writer.Header().Set("content-type", "application/xml")
    c.String(200, resp)
  } else {
    c.String(204, "")
  }
}

func list_sms_codes(c *gin.Context) {
  out := ""
  mutex.Lock()
  for _, entry := range entries {
    matches := code_regexp.FindStringSubmatch(entry.text)
    if len(matches) > 0 {
      out += matches[0] + "\n"
    } else {
      out += entry.text + "\n"
    }
  }
  mutex.Unlock()
  c.Writer.Header().Set("content-type", "text/plain")
  set_cors_headers(c)
  c.String(http.StatusOK, out)
}

func list_sms_full(c *gin.Context) {
  out := ""
  mutex.Lock()
  for _, entry := range entries {
    out += entry.text + "\n"
  }
  mutex.Unlock()
  c.Writer.Header().Set("content-type", "text/plain")
  c.String(http.StatusOK, out)
}
