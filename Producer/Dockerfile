FROM golang:alpine AS builder

WORKDIR /Producer

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /Producer/cmd

RUN go build -o main .

FROM alpine:latest 

WORKDIR /ProducerRoot/

COPY --from=builder /Producer/cmd/main .

COPY --from=builder /Producer/configs/config.yaml /ProducerRoot/configs/config.yaml

CMD [ "./main" ]