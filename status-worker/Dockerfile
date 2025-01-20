FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./main.go

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main .


######## New Stage #######


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/main .

EXPOSE 8080

CMD ["./main"]