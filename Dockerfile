FROM golang:1.16


WORKDIR /app/busycontainer

RUN mkdir -p cmd/busycontainer

COPY go.mod .
COPY go.sum .
COPY cmd/busycontainer/main.go cmd/busycontainer/

RUN go mod download
RUN go build -o out/busycontainer cmd/busycontainer/main.go 

EXPOSE 3998
EXPOSE 3999
EXPOSE 4000
EXPOSE 4001
EXPOSE 4002
EXPOSE 4003

CMD ["./out/busycontainer"]