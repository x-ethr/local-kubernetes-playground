FROM golang:1.22-alpine AS build

ARG SERVICE

ENV GOOS="linux"
ENV GOVCS="*:all"
ENV CGO_ENABLED="0"

ENV GOPRIVATE="github.com/iac-factory/*"

WORKDIR /

RUN apk add --no-cache git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . ./

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" go build --mod vendor --ldflags="-s -w -X 'main.sname=ETHR' -X 'main.version=$(head VERSION)' -X 'main.service=${SERVICE}'" -o /service

# --> Prevents shell access
RUN adduser -h "/dev/null" -g "" -s "/sbin/nologin" -D -H -u 10000 api-service-user

FROM scratch AS release

WORKDIR /

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /service /usr/local/bin/service
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER api-service-user

CMD ["service"]
