package main

import (
	"github.com/pkg/errors"
	"strings"
)

type Config struct {
	DataSet   string
	ConfigSet string
}

// ReplicaSet format <replicaSetName>/data1:27017,data2:27017,data3:27017
func ParseReplicaSet(definition string) (string, []string, error) {

	parts := strings.Split(definition, "/")
	if len(parts) != 2 {
		return "", nil, errors.New("Invalid ReplicaSet definition, expected <replicaSetName>/data1:27017,data2:27017,data3:27017")
	}

	replSetName := parts[0]

	members := strings.Split(parts[1], ",")

	if len(members) < 3 {
		return "", nil, errors.New("Invalid ReplicaSet definition, a minimum of 3 members required")
	}

	return replSetName, members, nil
}
