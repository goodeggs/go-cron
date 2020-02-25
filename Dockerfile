FROM golang:1.13.3

WORKDIR /go/src/go-cron
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["go-cron"]
