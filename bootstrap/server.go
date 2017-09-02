package main

import (
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type HttpServer struct {
	Config *Config
}

func (s *HttpServer) Start() {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		sets := make([]ReplicaSetStatus, 2)
		datars, err := GetReplicaSet(s.Config.DataSet)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cfgrs, err := GetReplicaSet(s.Config.ConfigSet)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sets[0] = *datars
		sets[1] = *cfgrs

		b, err := json.MarshalIndent(sets, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	logrus.Error(http.ListenAndServe(fmt.Sprintf(":%v", s.Config.Port), http.DefaultServeMux))
}

func GetReplicaSet(url string) (*ReplicaSetStatus, error) {
	parts := strings.Split(url, "/")
	mongoUrl := fmt.Sprintf("mongodb://%v/?replicaSet=%v", parts[1], parts[0])
	session, err := mgo.DialWithTimeout(mongoUrl, 5*time.Second)
	if err != nil {
		return nil, errors.Wrapf(err, "%v connection failed", url)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	status := &ReplicaSetStatus{}
	if err := session.Run("replSetGetStatus", &status); err != nil {
		return nil, errors.Wrap(err, "replSetGetStatus query failed")
	}

	return status, nil
}
