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
	"log"
	"time"
	"os"
	"strconv"
	"sync"
)
import "net/http"

var mutex *sync.Mutex

type entry struct {
	code string
	received_at int64
}

var entries []entry

func twiml(forward_number string, from_number string, code string) string {
twiml_template := `
<?xml version='1.0' encoding='UTF-8'?>
<Response>
    <Message to='%s'>[OTPBASE:%s] %s</Message>
</Response>
`
return fmt.Sprintf(twiml_template, forward_number, from_number, code)
}

var forward_number string
var ticker *time.Ticker

func expire() {
	for range ticker.C {
		mutex.Lock()
		for len(entries) > 0 {
			entry := entries[len(entries)-1]
			if time.Now().UnixNano() - entry.received_at > 60e9 {
				entries = entries[:len(entries)-1]
			} else {
				break
			}
		}
		mutex.Unlock()
	}
}

func add(c *gin.Context) {
	code := c.PostForm("Body")
	from_number := c.PostForm("From")

	if len(code) == 0 {
		//c.AbortWithError(400, errors.New("Empty body is not allowed"))
		c.String(400, "Empty body is not allowed")
		return
	}

	mutex.Lock()
	entries = append([]entry{entry{code, time.Now().UnixNano()}}, entries...)
	if len(entries) > 5 {
		entries = entries[:5]
	}
	mutex.Unlock()

	if forward_number != "" {
	resp := twiml(forward_number, from_number, code)
	c.Writer.Header().Set("content-type", "application/xml")
	c.String(200, resp)
	} else {
	c.String(204, "")
}
}

func list(c *gin.Context) {
	out := ""
	mutex.Lock()
	for _, entry := range entries {
		out += entry.code + "\n"
	}
	mutex.Unlock()
	c.Writer.Header().Set("content-type", "text/plain")
	c.String(http.StatusOK, out)
}

func main() {
	mutex = &sync.Mutex{}
	entries = make([]entry, 0)
	
	ticker = time.NewTicker(10*time.Second)
	go expire()

	// Disable Console Color
	// gin.DisableConsoleColor()

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	//router.Use(gin.Recovery())

	router.POST("/", add)
	router.GET("/", list)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	forward_number = os.Getenv("FORWARD")
	port := os.Getenv("PORT")
	var iport int
	var err error
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
