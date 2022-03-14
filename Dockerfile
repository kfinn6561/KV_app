FROM golang:1.17

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

WORKDIR /usr/src/app/main
RUN go build -v -o kv-store

EXPOSE 80

CMD ["/usr/src/app/main/kv-store", "--port=80"]