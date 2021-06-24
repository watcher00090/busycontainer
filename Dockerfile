FROM golang:1.10


WORKDIR /app/busycontainer

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o out/busycontainer cmd/busycontainer/main.go 

EXPOSE 3998
EXPOSE 3999
EXPOSE 4000
EXPOSE 4001
EXPOSE 4002
EXPOSE 4003

CMD ["./busycontainer"]