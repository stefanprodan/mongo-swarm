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
	flag.StringVar(&config.Mongos, "mongos", "", "Mongos list")
	flag.IntVar(&config.Retry, "retry", 100, "retry count")
	flag.IntVar(&config.Wait, "wait", 5, "wait time between retries in seconds")
	flag.IntVar(&config.Port, "port", 9090, "HTTP server port")
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

	err = dataReplSet.InitWithRetry(config.Retry, config.Wait)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("%v replica set initialized successfully", dataReplSetName)

	hasPrimary, err := dataReplSet.WaitForPrimary(config.Retry, config.Wait)
	if err != nil {
		logrus.Fatal(err)
	}
	dataReplSet.PrintStatus()
	if !hasPrimary {
		logrus.Fatalf("No primary node found for replica set %v", dataReplSetName)
	}

	cfgReplSetName, cfgMembers, err := ParseReplicaSet(config.ConfigSet)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Bootstrap started for config cluster %v members %v", cfgReplSetName, cfgMembers)

	cfgReplSet := &ReplicaSet{
		Name:    cfgReplSetName,
		Members: cfgMembers,
	}

	err = cfgReplSet.InitWithRetry(config.Retry, config.Wait)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("%v replica set initialized successfully", cfgReplSetName)

	hasPrimary, err = cfgReplSet.WaitForPrimary(config.Retry, config.Wait)
	if err != nil {
		logrus.Fatal(err)
	}
	cfgReplSet.PrintStatus()
	if !hasPrimary {
		logrus.Fatalf("No primary node found for replica set %v", cfgReplSetName)
	}

	mongosList, err := ParseMongos(config.Mongos)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Bootstrap started for mongos %v", mongosList)
	for _, mongos := range mongosList {

		err := pingWithRetry(mongos, config.Retry, config.Wait)
		if err != nil {
			logrus.Fatalf("% mongos connection failed", mongos)
		} else {
			logrus.Infof("%v is online", mongos)
		}

		m := &Mongos{
			Address:       mongos,
			ReplicaSetUrl: config.DataSet,
		}

		err = m.Init()
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("%v shard added", mongos)
	}

	logrus.Info("Bootstrap finished")

	logrus.Infof("Starting HTTP server on port %v", config.Port)
	server := HttpServer{
		Config: config,
	}
	go server.Start()

	//wait for exit signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-sigChan
	logrus.Infof("Shutting down %v signal received", sig)
}
