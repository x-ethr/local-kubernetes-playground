FROM golang:1.22-alpine as BUILD

ARG SERVICE

ENV GOOS "linux"
ENV GOVCS "*:all"
ENV CGO_ENABLED "0"

ENV GOPRIVATE "github.com/iac-factory/*"

WORKDIR /

RUN apk add --no-cache git bash

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . ./

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" go build --ldflags="-s -w -X 'main.header=ETHR' -X 'main.version=$(head VERSION)' -X 'main.service=${SERVICE}'" -o /service

# --> Prevents shell access
#RUN adduser -h "/dev/null" -g "" -s "/sbin/nologin" -D -H -u 10000 api-service-user
#
#FROM scratch as RELEASE
#
#WORKDIR /
#
#COPY --from=BUILD /etc/passwd /etc/passwd
#COPY --from=BUILD /service /usr/local/bin/service
#COPY --from=BUILD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# USER api-service-user

CMD ["/service"]
