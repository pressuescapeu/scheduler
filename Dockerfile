FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go env -w GOPROXY=https://proxy.golang.org,direct

RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scheduler-app ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/scheduler-app .

EXPOSE 8080

CMD ["./scheduler-app"]