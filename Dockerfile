FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY ./structs ./structs
COPY ./server ./server
RUN go mod download

WORKDIR /app/server

RUN go build -o server

FROM alpine:3
COPY --from=builder /app/server/server /bin/server
COPY --from=builder /app/server/.dbenv /bin/.dbenv
EXPOSE 8080
EXPOSE 5432

WORKDIR /bin

ENTRYPOINT ["./server"]
