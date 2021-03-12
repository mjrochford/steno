FROM golang:1.16 as build-env

WORKDIR /go/src/steno

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/steno

FROM scratch
COPY --from=build-env /go/bin/steno /go/bin/steno
ENTRYPOINT ["/go/bin/steno"]
