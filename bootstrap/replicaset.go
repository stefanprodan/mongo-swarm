package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ReplicaSet struct {
	Name    string
	Members []string
}

func ping(member string) error {

	session, err := mgo.DialWithTimeout(fmt.Sprintf(
		"%v?connect=direct", member), 5*time.Second)

	if err != nil {
		return errors.Wrap(err, "Connection failed")
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	err = session.Ping()
	return errors.Wrap(err, "Connection failed")
}

func pingWithRetry(member string, retry int, wait int) error {
	var err error
	for retry > 0 {
		err = ping(member)
		if err != nil {
			retry--
			logrus.Warnf("%v Retry %v after %v seconds", err.Error(), retry, wait)
			time.Sleep(time.Duration(wait) * time.Second)
		} else {
			return nil
		}
	}

	return errors.Wrapf(err, "%v ping failed", member)
}

func (r *ReplicaSet) init() error {

	session, err := mgo.DialWithTimeout(fmt.Sprintf(
		"%v?connect=direct", r.Members[0]), 5*time.Second)

	if err != nil {
		return errors.Wrap(err, "Connection failed")
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	memberList := make([]map[string]interface{}, len(r.Members))

	for i, item := range r.Members {
		memberList[i] = bson.M{"_id": i, "host": item}
	}

	config := bson.M{
		"_id":     r.Name,
		"members": memberList,
	}
	result := bson.M{}
	if err := session.Run(bson.M{"replSetInitiate": config}, &result); err != nil {
		return errors.Wrap(err, "replSetInitiate failed")
	}

	return nil
}

func (r *ReplicaSet) InitWithRetry(retry int, wait int) error {

	for _, member := range r.Members {
		err := pingWithRetry(member, retry, wait)
		if err != nil {
			return errors.Wrap(err, "ReplicaSet init failed")
		} else {
			logrus.Infof("%v member %v is online", r.Name, member)
		}
	}

	err := r.init()
	if err != nil {
		return errors.Wrap(err, "ReplicaSet init failed")
	}

	return nil
}
