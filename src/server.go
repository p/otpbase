package main

// An OTP in-memory database, meant to be used with a Twilio SMS receiver.
//
// Twilio should be configured to POST incoming SMS messages to /.
// They are handled by the `add` method, which parses the request,
// extracts the OTP code and stores it on top of the code pile.
// A GET request to / shows the most recent 5 codes received within
// the last 5 minutes.
//
// The program treats OTP codes as strings and as such will work with
// any data received via SMS.

// To forward OTP codes to another number in addition to recording them,
// set FORWARD environment variable to the phone number to forward to and
// set this program's URL as the webhook in Twilio; see
// https://stackoverflow.com/questions/40706883/forward-voice-call-and-invoke-webhook

import (
  //"errors"
  "fmt"
  "github.com/gin-gonic/gin"
  "os"
  "strconv"
  "sync"
  //"io"
  "io/ioutil"
  "log"
  "regexp"
  "strings"
  "time"

  bolt "go.etcd.io/bbolt"
  "html/template"
  "net/http"
)

var http_user, http_password string
var ticker *time.Ticker
var code_regexp *regexp.Regexp
var apps_template *template.Template
var db *bolt.DB

func add_app(c *gin.Context) {
  name := c.PostForm("name")
  secret := c.PostForm("secret")

  if len(name) == 0 {
    c.String(400, "Name is required")
    return
  }
  if len(secret) == 0 {
    c.String(400, "Secret is required")
    return
  }

  key, err := secret_to_key(secret)
  if err != nil {
    c.String(400, err.Error())
    return
  }

  err = db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("apps"))
    err := b.Put([]byte(name), key)
    return err
  })

  if err != nil {
    c.String(500, "Error saving: "+err.Error())
    return
  }

  c.Redirect(303, "/apps")
}

func delete_app(c *gin.Context) {
  name := c.Param("name")

  err := db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("apps"))
    err := b.Delete([]byte(name))
    return err
  })

  if err != nil {
    c.String(500, "Error removing: "+err.Error())
    return
  }

  c.Redirect(303, "/apps")
}

func apps(c *gin.Context) {
  m := make(map[string]string)
  db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("apps"))
    b.ForEach(func(k, v []byte) error {
      m[string(k)] = string(v)
      return nil
    })
    return nil
  })

  c.HTML(http.StatusOK, "/views/apps.html", gin.H{
    "apps": m,
  })
}

func app(c *gin.Context) {
  var key []byte
  name := c.Param("name")
  db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("apps"))
    key = b.Get([]byte(name))
    return nil
  })

  code, err := gen_totp(key)
  if err != nil {
    c.String(500, "Error generating: "+err.Error())
    return
  }

  set_cors_headers(c)
  c.String(200, code)
}

// loadTemplate loads templates embedded by go-assets-builder
func load_templates() (*template.Template, error) {
  t := template.New("")
  for name, file := range Assets.Files {
    if file.IsDir() || !strings.HasSuffix(name, ".html") {
      continue
    }
    h, err := ioutil.ReadAll(file)
    if err != nil {
      return nil, err
    }
    t, err = t.New(name).Parse(string(h))
    if err != nil {
      return nil, err
    }
  }
  return t, nil
}

func set_cors_headers(c *gin.Context) {
  c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
  c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
}

func main() {
  var err error

  mutex = &sync.Mutex{}
  entries = make([]entry, 0)
  code_regexp = regexp.MustCompile(CODE_REGEXP)

  http_user = os.Getenv("HTTP_USER")
  http_password = os.Getenv("HTTP_PASSWORD")
  if http_user == "" && http_password != "" {
    log.Fatal("HTTP_PASSWORD was specified but HTTP_USER was not, they need to be given together")
  }
  if http_user != "" && http_password == "" {
    log.Fatal("HTTP_USER was specified but HTTP_PASSWORD was not, they need to be given together")
  }

  apps_template, err = template.New("apps").Parse(`
  <b>Hello world</b>
`)
  if err != nil {
    //log.Fatal("Error loading apps template: " + err)
  }

  db_path := os.Getenv("DB_PATH")
  if db_path == "" {
    db_path = "otpbase.db"
  }
  db, err = bolt.Open(db_path, 0600, nil)
  if err != nil {
    log.Fatal("Error opening database")
  }
  defer db.Close()

  db.Update(func(tx *bolt.Tx) error {
    b, err := tx.CreateBucketIfNotExists([]byte("apps"))
    if err != nil {
      log.Fatal("Cannot create apps bucket")
    }
    b = b
    return nil
  })

  ticker = time.NewTicker(10 * time.Second)
  go expire_sms()

  // Disable Console Color
  // gin.DisableConsoleColor()

  debug := os.Getenv("DEBUG")
  if debug == "" {
    gin.SetMode(gin.ReleaseMode)
  }

  // Creates a gin router with default middleware:
  // logger and recovery (crash-free) middleware
  router := gin.Default()

  //router.LoadHTMLGlob("views/*.html")

  t, err := load_templates()
  if err != nil {
    panic(err)
  }
  router.SetHTMLTemplate(t)

  //router.Use(gin.Recovery())

  router.POST("/", receive_sms)

  var sms_router gin.IRouter
  if http_user != "" {
    authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
      http_user: http_password,
    }))
    sms_router = authorized
  } else {
    sms_router = router
  }
  sms_router.GET("/", list_sms_codes)
  sms_router.DELETE("/", clear_sms_codes)
  sms_router.GET("/full", list_sms_full)
  router.GET("/apps", apps)
  router.GET("/apps/:name", app)
  router.POST("/apps/:name/delete", delete_app)
  router.POST("/apps", add_app)
  router.GET("/robots.txt", robots_txt)

  // By default it serves on :8080 unless a
  // PORT environment variable was defined.
  forward_number = os.Getenv("FORWARD")
  port := os.Getenv("PORT")
  var iport int
  if port == "" {
    iport = 8092
  } else {
    iport, err = strconv.Atoi(port)
    if err != nil {
      log.Fatal(err)
    }
  }
  router.Run(fmt.Sprintf(":%d", iport))
  // router.Run(":3000") for a hard coded port
}
