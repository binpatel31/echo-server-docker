FROM golang:1.16 AS build

ENV GOPROXY="direct"

RUN mkdir /tmpapp && mkdir /tmpapp/bin
WORKDIR /tmpapp
# NOTE: relative path = /tmpapp
COPY *.yaml *.go go.mod go.sum ./
COPY templates ./templates
COPY cmd ./cmd
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /tmpapp/bin/echoserver

FROM scratch AS final

COPY --from=build /tmpapp/bin/echoserver /echoserver
EXPOSE 8080
ENTRYPOINT ["./echoserver"]
