FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./main.go ./

RUN go build main.go

CMD ["./main"]
