FROM  golang:1.21-alpine as build

COPY . /go/src/github.com/azhuravlev/docker-nats-logger-plugin
RUN cd /go/src/github.com/azhuravlev/docker-nats-logger-plugin && \
    go get && \
    go build --ldflags '-extldflags "-static"' -o /usr/bin/docker-nats-logger-plugin

FROM alpine

RUN mkdir -p /run/docker/plugins
COPY --from=build /usr/bin/docker-nats-logger-plugin /usr/bin/docker-nats-logger-plugin
CMD ["docker-nats-logger-plugin"]
