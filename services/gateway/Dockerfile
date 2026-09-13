FROM golang:1.26-alpine AS build
WORKDIR /app
COPY contracts/gen/go/ contracts/gen/go/
COPY services/gateway/ services/gateway/
WORKDIR /app/services/gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/server

FROM alpine:3.23
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=build /out/gateway /usr/local/bin/gateway
EXPOSE 8080
ENTRYPOINT ["gateway"]
