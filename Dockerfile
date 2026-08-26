# Build stage
FROM golang:1.26-alpine3.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
RUN ls -la /go/bin

# Run stage
FROM alpine:3.24
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /go/bin/migrate /app/migrate
COPY app.env .
COPY start.sh .
COPY db/migration ./migration

EXPOSE 8080
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
