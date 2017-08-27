FROM golang:1.8.3-alpine

# copy code
ADD . /go/src/github.com/stefanprodan/mongo-swarm/boostrap

# solution root
WORKDIR /go/src/github.com/stefanprodan/mongo-swarm/boostrap

# pull deps
RUN apk add --no-cache --virtual git
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure

# output
RUN mkdir /go/dist
VOLUME /go/dist
