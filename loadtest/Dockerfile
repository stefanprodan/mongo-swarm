FROM golang:1.8.3-alpine as builder

ARG APP_VERSION=unkown

# copy code
ADD . /go/src/github.com/stefanprodan/mongo-swarm/loadtest

# solution root
WORKDIR /go/src/github.com/stefanprodan/mongo-swarm/loadtest

# pull deps
RUN apk add --no-cache --virtual git
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure

# output
RUN mkdir /go/dist
RUN go build -ldflags "-X main.version=$APP_VERSION" \
    -o /go/dist/loadtest github.com/stefanprodan/mongo-swarm/loadtest

FROM alpine:latest

COPY --from=builder /go/dist/loadtest /mongo-swarm/loadtest

RUN chmod 777 /mongo-swarm/loadtest

EXPOSE 9999
WORKDIR /mongo-swarm
ENTRYPOINT ["/mongo-swarm/loadtest"]
