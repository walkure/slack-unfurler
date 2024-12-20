FROM golang:1.23.3-alpine3.19 AS builder

WORKDIR /app
COPY . /app/
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && apk del tzdata
RUN go mod download
RUN addgroup -g 6128 -S nonroot && adduser -u 6128 -S nonroot -G nonroot
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/socket ./socket/ 

FROM scratch AS runner

USER nonroot

ENV TZ=Asia/Tokyo
COPY --from=builder /etc/localtime /usr/share/zoneinfo/Asia/Tokyo

COPY --from=builder /etc/passwd /etc/group /etc/
COPY --from=builder --chown=nonroot:nonroot /bin/socket /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/app/socket"]
