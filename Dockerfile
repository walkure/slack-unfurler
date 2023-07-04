FROM golang:1.19.1-alpine3.16 as builder

WORKDIR /app
COPY . /app/
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && apk del tzdata
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/socket ./socket/ 

FROM busybox:1.35.0-musl as runner

ENV TZ=Asia/Tokyo
COPY --from=builder  /bin/socket /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT /app/socket
