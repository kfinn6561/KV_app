FROM golang:1.17

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o kv-store

EXPOSE 80

CMD ["./kv-store", "--port=80"]