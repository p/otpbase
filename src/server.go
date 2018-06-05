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

import (
	//"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)
import "net/http"

var codes []string

func add(c *gin.Context) {
	code := c.PostForm("code")
	
	codes = append([]string{code}, codes...)
	if len(codes) > 5 {
		codes = codes[:5]
		}
	
	c.JSON(http.StatusOK, gin.H{})
}

func list(c *gin.Context) {
out := ""
for _,code := range codes {
out += code + "\n"
}
	c.String(http.StatusOK, out)
}

func main() {
	codes = make([]string, 0)

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
