FROM golang:1.23-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
ENTRYPOINT ["./main"]