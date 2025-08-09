FROM golang:latest AS builder
LABEL author="gqvz"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN $(go env GOPATH)/bin/swag init -g cmd/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /main ./cmd/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/database ./database
COPY --from=builder /main .
COPY --from=builder /app/docs ./docs

CMD ["/app/main"]