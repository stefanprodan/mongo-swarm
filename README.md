# mongo-swarm

[![Build Status](https://travis-ci.org/stefanprodan/mongo-swarm.svg?branch=master)](https://travis-ci.org/stefanprodan/mongo-swarm)

Setup MongoDB sharded clusters on Docker Swarm

![Overview](https://github.com/stefanprodan/mongo-swarm/blob/master/diagrams/rsz_1mongo-swarm.png)

### Prerequisites 

In oder to deploy the MongoDB Cluster you need to have a Docker Swarm cluster made out of eleven nodes:

* 3 Swarm manager nodes (prod-manager-1, prod-manager-2, prod-manager-3)
* 3 Mongo data nodes (prod-mongodata-1, prod-mongodata-2, prod-mongodata-3)
* 3 Mongo config nodes (prod-mongocfg-1, prod-mongocfg-2, prod-mongocfg-3)
* 2 Mongo router nodes (prod-mongos-1, prod-mongos-2)

You can name your Swarm nodes however you want, 
the bootstrap process uses placement restrictions based on nodes labels. For the bootstraping to take place 
you need to apply the following labels:

**Mongo data nodes**

```bash
docker node update --label-add mongo.role=data1 prod-mongodata-1
docker node update --label-add mongo.role=data2 prod-mongodata-2
docker node update --label-add mongo.role=data3 prod-mongodata-3
```

**Mongo config nodes**

```bash
docker node update --label-add mongo.role=cfg1 prod-mongocfg-1
docker node update --label-add mongo.role=cfg2 prod-mongocfg-2
docker node update --label-add mongo.role=cfg3 prod-mongocfg-3
```

**Mongos nodes**

```bash
docker node update --label-add mongo.role=mongos1 prod-mongos-1
docker node update --label-add mongo.role=mongos2 prod-mongos-2
```

### Cluster bootstrap

Clone this repository and run bootstrap.sh on a Swarm manager node:

```bash
$ git clone https://github.com/stefanprodan/mongo-swarm
$ cd mongo-swarm

$ ./bootstrap.sh
```

The bootstrap.sh script creates two overlay networks and deploys the mongo stack:

```bash
docker network create --attachable -d overlay mongo
docker network create --attachable -d overlay mongos

docker stack deploy -c swarm-compose.yml mongo
```

The config and data replica sets are isolated from the rest of the swarm in the mongo overlay network. The 
routers, Mongos1 and Mongos2 are connected to the mongo network and to the mongos network. You should attach 
application containers to the mongos network in order to communicate with the MongoDB Cluster.

After the stack has been deploy the mongo-bootstrap container will do the following:

* waits for the data nodes to be online
* joins the data nodes into a replica set (datars)
* waits for the config nodes to be online
* joins the config nodes into a replica set (cfgrs)
* waits for the mongos nodes to be online
* adds the data replica set shard to the mongos instances

In order to monitor the bootstrap process you can watch the mongo-bootstrap service logs:

```bash
docker service logs mongo_bootstrap

msg="Bootstrap started for data cluster datars members [data1:27017 data2:27017 data3:27017]"
msg="datars member data1:27017 is online"
msg="datars member data2:27017 is online"
msg="datars member data3:27017 is online"
msg="datars replica set initialized successfully"
msg="datars member data1:27017 state PRIMARY"
msg="datars member data2:27017 state SECONDARY"
msg="datars member data3:27017 state SECONDARY"
msg="Bootstrap started for config cluster cfgrs members [cfg1:27017 cfg2:27017 cfg3:27017]"
msg="cfgrs member cfg1:27017 is online"
msg="cfgrs member cfg2:27017 is online"
msg="cfgrs member cfg3:27017 is online"
msg="cfgrs replica set initialized successfully"
msg="cfgrs member cfg1:27017 state PRIMARY"
msg="cfgrs member cfg2:27017 state SECONDARY"
msg="cfgrs member cfg3:27017 state SECONDARY"
msg="Bootstrap started for mongos [mongos1:27017 mongos2:27017]"
msg="mongos1:27017 is online"
msg="mongos1:27017 shard added"
msg="mongos2:27017 is online"
msg="mongos2:27017 shard added"
```
