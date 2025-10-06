package main

import (
	"flag"
	"fmt"
	"goProxy/core/config"
	"goProxy/core/pnc"
	"goProxy/core/proxy"
	"goProxy/core/server"
	"io"
	"log"
	"os"
	"time"
)

var Fingerprint string = "S3LF_BU1LD_0R_M0D1F13D" // 455b9300-0a6f-48f1-82ee-bb1f6cf43500

func main() {

	proxy.Fingerprint = Fingerprint

	// Daemon mode
	daemon := flag.Bool("daemon", false, "run as daemon")
	dFlag := flag.Bool("d", false, "run as daemon (shorthand)")
	flag.Parse()
	if *dFlag || *daemon {
		proxy.DisableMonitor = true
	}

	logFile, err := os.OpenFile("crash.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	pnc.InitHndl()

	defer pnc.PanicHndl()

	//Disable Error Logging
	log.SetOutput(io.Discard /*logFile*/) // if we ever need to log to a file

	fmt.Println("Starting Proxy ...")

	config.Load()
	if err := config.LoadIpWhitelist(); err != nil {
		fmt.Println("Error while loading whitelist: " + err.Error())
	}

	fmt.Println("Loaded Config ...")

	// Wait for everything to be initialised
	fmt.Println("Initialising ...")
	go server.Monitor()
	
	// Optimized initialization wait with timeout
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for !proxy.Initialised {
		select {
		case <-ticker.C:
			// Continue waiting
		case <-timeout:
			log.Fatal("Initialization timeout - proxy failed to initialize within 30 seconds")
		}
	}

	go server.Serve()

	//Keep server running
	select {}
}
