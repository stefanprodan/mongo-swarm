package main

import (
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"time"
)

type Repository struct {
	Session *mgo.Session
}

func NewRepository(uri string) (*Repository, error) {
	session, err := mgo.DialWithTimeout(uri, 5*time.Second)
	if err != nil {
		return nil, errors.Wrapf(err, "%v connection failed", uri)
	}

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("test").C("log")
	err = c.EnsureIndexKey("timestamp")
	if err != nil {
		return nil, errors.Wrapf(err, "%v index creation failed", uri)
	}

	repo := &Repository{
		Session: session,
	}
	return repo, nil
}
