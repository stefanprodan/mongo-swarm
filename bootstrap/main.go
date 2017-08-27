package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

var version = "undefined"

func main() {
	var config = &Config{}
	flag.StringVar(&config.DataSet, "dataSet", "", "MongoDB data cluster")
	flag.StringVar(&config.ConfigSet, "configSet", "", "MongoDB config cluster")
	appVersion := flag.Bool("v", false, "prints version")
	flag.Parse()

	if *appVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	//wait for exit signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-sigChan
	logrus.Infof("Shutting down %v signal received", sig)
}
