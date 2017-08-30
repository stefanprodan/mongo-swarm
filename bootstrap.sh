#!/usr/bin/env bash

docker network create --attachable -d overlay mongo
docker network create --attachable -d overlay mongos

docker stack deploy -c swarm-compose.yml mongo
