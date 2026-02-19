FROM golang:1.25 AS builder

WORKDIR /app

COPY /go.mod /go.sum ./

RUN go mod download

COPY ./app .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

FROM alpine:latest AS final

WORKDIR /app

COPY --from=builder ./app/main .

CMD [ "./main" ]