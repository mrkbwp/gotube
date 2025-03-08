FROM golang:1.22.10-alpine

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /migrations

COPY *.sql .

COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]