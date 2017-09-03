package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type HttpServer struct {
	Port       int
	Repository *Repository
}

func (s *HttpServer) Start() {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		session := s.Repository.Session.Copy()
		defer session.Close()

		log := &AccessLog{
			Timestamp: time.Now().UTC(),
			UserAgent: string(req.Header.Get("User-Agent")),
		}

		c := session.DB("test").C("log")

		err := c.Insert(log)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logs := []AccessLog{}

		err = c.Find(nil).Sort("-timestamp").Limit(10).All(&logs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.MarshalIndent(logs, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.Port), http.DefaultServeMux))
}

type AccessLog struct {
	UserAgent string    `json:"ua" bson:"ua"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
