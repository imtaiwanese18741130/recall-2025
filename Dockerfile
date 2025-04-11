FROM golang:1.23-alpine AS golang-builder

WORKDIR /go/src/github.com/imtaiwanese18741130/recall-2025
COPY . .
RUN go mod tidy
RUN go build -o /go/bin/recall-2025 .

FROM alpine:latest
COPY --from=golang-builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=golang-builder /go/bin/recall-2025 /var/www/app/
COPY --from=golang-builder /go/src/github.com/imtaiwanese18741130/recall-2025/json-config /var/www/app/json-config
COPY --from=golang-builder /go/src/github.com/imtaiwanese18741130/recall-2025/assets /var/www/app/assets
COPY --from=golang-builder /go/src/github.com/imtaiwanese18741130/recall-2025/templates /var/www/app/templates
WORKDIR /var/www/app

EXPOSE 8964
ENV APP_PORT 8964

ENTRYPOINT ["./recall-2025"]
