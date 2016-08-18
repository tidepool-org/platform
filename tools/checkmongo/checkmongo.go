package main

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("ERROR: Specify Mongo server address(es), database, and SSL(true|false)")
		os.Exit(1)
	}

	addresses := strings.Split(os.Args[1], ",")
	database := os.Args[2]
	ssl := os.Args[3]

	dialInfo := &mgo.DialInfo{
		Addrs:    addresses,
		Timeout:  10 * time.Second,
		Database: database,
	}

	if ssl == "true" {
		dialInfo.DialServer = func(serverAddr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", serverAddr.String(), &tls.Config{InsecureSkipVerify: true})
		}
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Printf("ERROR: Failure connecting to Mongo: %s\n", err.Error())
		os.Exit(1)
	}

	if err = session.Ping(); err != nil {
		fmt.Printf("ERROR: Failure during Ping: %s\n", err.Error())
		os.Exit(1)
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		fmt.Printf("ERROR: Failure obtaining BuildInfo: %s\n", err.Error())
		os.Exit(1)
	}

	mode := session.Mode()
	safe := session.Safe()
	liveServers := session.LiveServers()

	fmt.Printf("BuildInfo: %#v\n", buildInfo)
	fmt.Printf("Mode: %#v\n", mode)
	fmt.Printf("Safe: %#v\n", safe)
	fmt.Printf("LiveServers: %#v\n", liveServers)

	os.Exit(0)
}
