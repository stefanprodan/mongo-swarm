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

func (r *ReplicaSet) init() error {
	session, err := mgo.DialWithTimeout(fmt.Sprintf(
		"%v?connect=direct", r.Members[0]), 5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "%v connection failed", r.Members[0])
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
		if err.Error() == "already initialized" {
			logrus.Warnf("%v replica set already initialized", r.Name)
		} else {
			return errors.Wrapf(err, "%v replSetInitiate failed", r.Name)
		}
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

func (r *ReplicaSet) hasPrimary() (bool, error) {
	session, err := mgo.DialWithTimeout(fmt.Sprintf(
		"%v?connect=direct", r.Members[0]), 5*time.Second)
	if err != nil {
		return false, errors.Wrapf(err, "%v connection failed", r.Members[0])
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	status := &ReplicaSetStatus{}
	if err := session.Run("replSetGetStatus", &status); err != nil {
		return false, errors.Wrapf(err, "%v replSetGetStatus failed", r.Name)
	} else {
		for _, m := range status.Members {
			if m.StateStr == "PRIMARY" {
				return true, nil
			}
		}
	}

	return false, nil
}

func (r *ReplicaSet) WaitForPrimary(retry int, wait int) (bool, error) {
	var hasPrimary bool
	var err error
	for retry > 0 {
		hasPrimary, err = r.hasPrimary()
		if err != nil {
			return false, errors.Wrapf(err, "connection failed")
		}
		if hasPrimary {
			return true, nil
		}
		retry--
		logrus.Warnf("No primary detected for set %v retying in %v seconds", r.Name, wait)
		time.Sleep(time.Duration(wait) * time.Second)
	}

	return hasPrimary, nil
}

func (r *ReplicaSet) PrintStatus() error {
	session, err := mgo.DialWithTimeout(fmt.Sprintf(
		"%v?connect=direct", r.Members[0]), 5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "%v connection failed", r.Members[0])
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	status := &ReplicaSetStatus{}
	if err := session.Run("replSetGetStatus", &status); err != nil {
		return errors.Wrapf(err, "%v replSetGetStatus failed", r.Name)
	} else {
		for _, m := range status.Members {
			logrus.Infof("%v member %v state %v", status.Name, m.Name, m.StateStr)
			if len(m.ErrMsg) > 0 {
				logrus.Warnf("%v member %v error %v", status.Name, m.Name, m.ErrMsg)
			}
		}
	}

	return nil
}

// replica set replSetGetStatus response object
type ReplicaSetStatus struct {
	Name    string                   `bson:"set" json:"set"`
	Members []ReplicaSetMemberStatus `bson:"members" json:"members"`
}

// replica set member replSetGetStatus response object
type ReplicaSetMemberStatus struct {
	Id       int           `bson:"_id" json:"id"`
	Name     string        `bson:"name" json:"name"`
	ErrMsg   string        `bson:"errmsg" json:"errmsg"`
	Healthy  bool          `bson:"health" json:"healthy"`
	StateStr string        `bson:"stateStr" json:"state"`
	Uptime   time.Duration `bson:"uptime" json:"uptime"`
}
