FROM golang:1.8.3-alpine

COPY /dist/boostrap /mongo-swarm/boostrap

RUN chmod 777 /mongo-swarm/boostrap


WORKDIR /mongo-swarm
ENTRYPOINT ["/mongo-swarm/boostrap"]
