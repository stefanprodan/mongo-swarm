package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

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
			logrus.Warnf("%v is offline retrying in %v seconds", member, wait)
			time.Sleep(time.Duration(wait) * time.Second)
		} else {
			return nil
		}
	}

	return errors.Wrapf(err, "%v ping failed", member)
}
