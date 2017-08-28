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
	flag.IntVar(&config.Retry, "retry", 100, "retry count")
	appVersion := flag.Bool("v", false, "prints version")
	flag.Parse()

	if *appVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	dataReplSetName, dataMembers, err := ParseReplicaSet(config.DataSet)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Bootstrap started for data cluster %v members %v", dataReplSetName, dataMembers)

	dataReplSet := &ReplicaSet{
		Name:    dataReplSetName,
		Members: dataMembers,
	}

	err = dataReplSet.InitWithRetry(config.Retry, 1)
	if err != nil {
		logrus.Fatal(err)
	}

	dataReplSet.PrintStatus()

	cfgReplSetName, cfgMembers, err := ParseReplicaSet(config.ConfigSet)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Bootstrap started for config cluster %v members %v", cfgReplSetName, cfgMembers)

	cfgReplSet := &ReplicaSet{
		Name:    cfgReplSetName,
		Members: cfgMembers,
	}

	err = cfgReplSet.InitWithRetry(config.Retry, 1)
	if err != nil {
		logrus.Fatal(err)
	}

	cfgReplSet.PrintStatus()

	//wait for exit signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-sigChan
	logrus.Infof("Shutting down %v signal received", sig)
}
