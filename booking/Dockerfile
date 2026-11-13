# The tests run as part of the build, so an image only exists for code that
# passes them.
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
ENV CGO_ENABLED=0
RUN go vet ./... && go test ./... && go build -trimpath -ldflags="-s -w" -o /booking .

FROM alpine:3.20
COPY --from=build /booking /usr/local/bin/booking
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/booking"]
