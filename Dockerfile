FROM golang:1.22.2-alpine

RUN apk add make

COPY . /app

WORKDIR /app

RUN make build-linux

CMD ["./bin/app"]
