FROM golang:latest
LABEL authors="gqvz"
WORKDIR /app

RUN apt-get update && apt-get install -y make

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make swagger

CMD ["make", "run"]