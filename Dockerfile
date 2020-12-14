FROM golang:1.15.6-buster AS build
WORKDIR /src
COPY . .
RUN make build
ENTRYPOINT ["/src/bin//opentelemetry-go-http-sample"]
