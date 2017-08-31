# mongo-swarm

[![Build Status](https://travis-ci.org/stefanprodan/mongo-swarm.svg?branch=master)](https://travis-ci.org/stefanprodan/mongo-swarm)
[![Docker Image](https://images.microbadger.com/badges/image/stefanprodan/mongo-bootstrap.svg)](https://hub.docker.com/r/stefanprodan/mongo-bootstrap)

Mongo-swarm is a POC project that automates the bootstrapping process of a MongoDB cluster for production use.
With a single command you can deploy the _Mongos_, _Config_ and _Data_ replica sets onto Docker Swarm, 
forming a high-available MongoDB cluster capable of surviving multiple nodes failure without service interruption. 
The Docker stack is composed of two MongoDB replica sets, two Mongos instances and the
mongo-bootstrap service. Mongo-bootstrap is written in Go and handles the replication, sharding and 
routing configuration.

![Overview](https://github.com/stefanprodan/mongo-swarm/blob/master/diagrams/mongo-swarm.png)

### Prerequisites 

In oder to deploy the MongoDB stack you should have a Docker Swarm cluster made out of eleven nodes:

* 3 Swarm manager nodes (prod-manager-1, prod-manager-2, prod-manager-3)
* 3 Mongo data nodes (prod-mongodata-1, prod-mongodata-2, prod-mongodata-3)
* 3 Mongo config nodes (prod-mongocfg-1, prod-mongocfg-2, prod-mongocfg-3)
* 2 Mongo router nodes (prod-mongos-1, prod-mongos-2)

You can name your Swarm nodes however you want, 
the bootstrap process uses placement restrictions based on nodes labels. 
For the bootstrapping to take place you need to apply the following labels:

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

### Deploy

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

**Networking**

The config and data replica sets are isolated from the rest of the swarm in the `mongo` overlay network. 
The routers, Mongos1 and Mongos2 are connected to the `mongo` network and to the `mongos` network. 
You should attach application containers to the `mongos` network in order to communicate with 
the MongoDB Cluster.

**Persistent storage** 

At first run, each data and config node will be provisioned with a named Docker volume. This 
ensures the MongoDB databases will not be purged if you restart or update the MongoDB cluster. Even if you 
remove the whole stack the volumes will remain on the disk. If you want to delete the MongoDB data and config 
you have to run `docker volume purge` on each Swarm node.

**Bootstrapping**

After the stack has been deploy the mongo-bootstrap container will do the following:

* waits for the data nodes to be online
* joins the data nodes into a replica set (datars)
* waits for the config nodes to be online
* joins the config nodes into a replica set (cfgrs)
* waits for the mongos nodes to be online
* adds the data replica set shard to the mongos instances

You can monitor the bootstrap process by watching the mongo-bootstrap service logs:

```bash
$ docker service logs -f mongo_bootstrap

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

**High availability**

A MongoDB cluster provisioned with mongo-swarm can survive node failures and will 
start an automatic failover if:

* the primary data node goes down
* the primary config node goes down
* one of the mongos nodes goes down

When the primary data or config nod goes down, the Mongos instances will detect the new 
primary node and will reroute all the traffic to it. If a Mongos node goes down and your applications are 
configured to use both Mongos nodes, the Mongo driver will switch to the online Mongos instance. When you 
recover a failed data or config nod, this node will rejoin the replica set and resync if the oplog size allows it.

If you want the cluster to outstand more than one node failure per replica set, you can 
horizontally scale up the data and config sets by modifying the swarm-compose.yml file. 
Always have an odd number of nodes per replica set to avoid split brain situations. 

**Client connectivity**

To test the Mongos connectivity you can run an interactive mongo container attached to the mongos network:

```bash
$ docker run --network mongos -it mongo:3.4 mongo mongos1:27017 

mongos> use test
switched to db test

mongos> db.demo.insert({text: "demo"})
WriteResult({ "nInserted" : 1 })

mongos> db.demo.find()
{ "_id" : ObjectId("59a6fa01e33a5cec9872664f"), "text" : "demo" }
```

The Mongo clients should connect to all Mongos nodes that are running on the mongos overlay network. 
Here is an example with the mgo golang MongoDB driver:

```go
session, err := mgo.Dial("mongodb://mongos1:27017,mongos2:27017/")
```

**Local deployment**

If you want to run the MongoDB cluster on a single Docker machine without Docker Swarm mode you can use 
the local compose file. I use it for debugging on Docker for Mac.

```bash
$ docker-compose -f local-compose.yml up -d
``` 

This will start all the MongoDB services and mongo-bootstrap on the bridge network without persistent storage. 
  
