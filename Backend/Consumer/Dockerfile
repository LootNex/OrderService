FROM golang:alpine AS builder

WORKDIR /Consumer

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /Consumer/cmd

RUN go build -o main .

FROM alpine:latest

WORKDIR /Consumer/

COPY --from=builder /Consumer/cmd/main .

COPY --from=builder /Consumer/configs/config.yaml /Consumer/configs/config.yaml

COPY --from=builder /Consumer/internal/db/postgresql/migrations /Consumer/internal/db/postgresql/migrations

EXPOSE 8081

CMD [ "./main" ]