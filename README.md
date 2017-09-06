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
the bootstrap process uses placement restrictions based on the `mongo.role` label. 
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

Clone this repository and run the bootstrap script on a Docker Swarm manager node:

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

When the primary data or config node goes down, the Mongos instances will detect the new 
primary node and will reroute all the traffic to it. If a Mongos node goes down and your applications are 
configured to use both Mongos nodes, the Mongo driver will switch to the online Mongos instance. When you 
recover a failed data or config node, this node will rejoin the replica set and resync if the oplog size allows it.

If you want the cluster to outstand more than one node failure per replica set, you can 
horizontally scale up the data and config sets by modifying the swarm-compose.yml file. 
Always have an odd number of nodes per replica set to avoid split brain situations. 

You can test the automatic failover by killing or removing the primary data and config nodes:

```bash
root@prod-data1-1:~# docker kill mongo_data1.1....
root@prod-cfg1-1:~# docker rm -f mongo_cfg1.1....
```

When you bring down the two instances Docker Swarm will start new containers to replace the killed ones. 
The data and config replica sets will choose a new leader and the newly started instances will join the 
cluster as followers. 

You can check the cluster state by doing an HTTP GET on mongo-bootstrap port 9090.

```json
docker run --rm --network mongo tutum/curl:alpine curl bootstrap:9090
```

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

**Load testing**

You can run load tests for the MongoDB cluster using the loadtest app. 

Start 3 loadtest instances on the mongos network:

```bash
docker stack deploy -c swarm-loadtest.yml lt
``` 

The loadtest app is a Go web service that connects to the two Mongos nodes and does an insert and select:

```go
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
```

The loadtest service is exposed on the internet on port 9999. 
You can run the load test using rakyll/hey or Apache bench.

```bash
#install hey
go get -u github.com/rakyll/hey

#do 10K requests 
hey -n 10000 -c 100 -m GET http://<SWARM-PUBLIC-IP>:9999/
```

While running the load test you could kill a _Mongos_, _Data_ and _Config_ node and see 
what's the failover impact.

Running the load test with a single loadtest instance:

```bash
Summary:
  Total:	58.3945 secs
  Slowest:	2.5077 secs
  Fastest:	0.0588 secs
  Average:	0.5608 secs
  Requests/sec:	171.2490
  Total data:	8508290 bytes
  Size/request:	850 bytes

Response time histogram:
  0.304 [1835]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.549 [3781]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.793 [2568]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  1.038 [1153]	|∎∎∎∎∎∎∎∎∎∎∎∎
  1.283 [400]	|∎∎∎∎
```

Running the load test with 3 loadtest instances:

```bash
Summary:
  Total:	35.5129 secs
  Slowest:	1.9471 secs
  Fastest:	0.0494 secs
  Average:	0.3223 secs
  Requests/sec:	281.5877
  Total data:	8508392 bytes
  Size/request:	850 bytes

Response time histogram:
  0.239 [5040]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.429 [2358]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.619 [1235]	|∎∎∎∎∎∎∎∎∎∎
  0.808 [741]	|∎∎∎∎∎∎
  0.998 [396]	|∎∎∎
```

Scaling up the application from one instance to three instances made the load test 23 seconds faster and the 
requests per second rate went from 171 to 281.

**Monitoring with Weave Scope**

Monitoring the load test with Weave Cloud shows how the traffic is being routed by the Docker Swarm 
load balancer and by the Mongos instances:

![Traffic](https://github.com/stefanprodan/mongo-swarm/blob/master/diagrams/weave-scope.png)

Weave Scope is a great tool for visualising network traffic between containers and/or Docker Swarm nodes. 
Besides traffic you can also monitor system load, CPU and memory usage. Recording multiple
load test sessions with Scope you can determine what's the maximum load your infrastructure can take without
a performance degradation. 

Monitoring a Docker Swarm cluster with Weave Cloud is as simple as deploying a Scope container on each Swarm node. 
More info on installing Weave Scope with Docker can be found [here](https://www.weave.works/docs/scope/latest/installing/).



**Local deployment**

If you want to run the MongoDB cluster on a single Docker machine without Docker Swarm mode you can use 
the local compose file. I use it for debugging on Docker for Mac.

```bash
$ docker-compose -f local-compose.yml up -d
``` 

This will start all the MongoDB services and mongo-bootstrap on the bridge network without persistent storage. 
