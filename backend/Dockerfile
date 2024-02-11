FROM golang:1.21-alpine

WORKDIR /backend

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /cmd/main.go

CMD ["./main"]