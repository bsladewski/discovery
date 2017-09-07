// The discovery command starts a discovery server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bsladewski/discovery"
)

// main Starts a discovery server as specified on the command line.
func main() {
	// parse command line arguments
	portPtr := flag.Int("port", 80, "specifies the port this server should use")
	logPtr := flag.String("log", "", "specifies destination file for logging")
	userPtr := flag.String("user", "", "specifies username for basic auth")
	passPtr := flag.String("pass", "", "specifies password for basic auth")
	certPtr := flag.String("cert", "", "specifies TLS certificate file")
	keyPtr := flag.String("key", "", "specifies TLS key file")
	flag.Parse()
	if (*certPtr != "" && *keyPtr == "") || (*certPtr == "" && *keyPtr != "") {
		fmt.Fprintf(os.Stderr, "TLS requires both certificate and key!\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if (*userPtr != "" && *passPtr == "") || (*userPtr == "" && *passPtr != "") {
		fmt.Fprintf(os.Stderr, "Basic auth requires both user and password!\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// configure logging
	if *logPtr != "" {
		file, err := os.OpenFile(*logPtr, os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logging: %s\n",
				err.Error())
			os.Exit(1)
		}
		log.SetOutput(file)
	}

	// run the server
	log.Printf("starting service on port %d!\n", *portPtr)
	auth := discovery.NullAuthenticator
	if *userPtr != "" {
		auth = discovery.NewBasicAuthenticator(*userPtr, *passPtr)
	}
	server := discovery.NewRandomServer(*portPtr, auth)
	var err error
	if *certPtr != "" {
		err = server.ListenAndServe()
	} else {
		err = server.ListenAndServeTLS(*certPtr, *keyPtr)
	}
	log.Printf("stopping service on port %d: %s\n", *portPtr, err.Error())
}
