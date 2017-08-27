FROM alpine:latest

COPY /dist/bootstrap /mongo-swarm/bootstrap

RUN chmod 777 /mongo-swarm/bootstrap


WORKDIR /mongo-swarm
ENTRYPOINT ["/mongo-swarm/bootstrap"]
