package main

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
		fmt.Println("ERROR: Specify Mongo server address(es), database, and TLS(true|false)")
		os.Exit(1)
	}

	addresses := strings.Split(os.Args[1], ",")
	database := os.Args[2]
	enableTLS := os.Args[3]

	dialInfo := &mgo.DialInfo{
		Addrs:    addresses,
		Timeout:  10 * time.Second,
		Database: database,
	}

	if enableTLS == "true" {
		dialInfo.DialServer = func(serverAddr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", serverAddr.String(), &tls.Config{InsecureSkipVerify: true})
		}
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Println("ERROR: Unable to connect to Mongo:", err)
		os.Exit(1)
	}

	if err = session.Ping(); err != nil {
		fmt.Println("ERROR: Unable to perform Ping:", err)
		os.Exit(1)
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		fmt.Println("ERROR: Unable to obtain BuildInfo:", err)
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
