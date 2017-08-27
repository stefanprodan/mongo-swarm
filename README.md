# mongo-swarm

Setup MongoDB shared clusters on Docker Swarm

### Swarm setup

Swarm nodes list:

* 3 Swarm manager nodes (prod-manager-1, prod-manager-2, prod-manager-3)
* 3 Mongo data nodes (prod-mongodata-1, prod-mongodata-2, prod-mongodata-3)
* 3 Mongo config nodes (prod-mongocfg-1, prod-mongocfg-2, prod-mongocfg-3)


Swarm nodes labels:

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

