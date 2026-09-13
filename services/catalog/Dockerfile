FROM golang:1.26-alpine AS build
WORKDIR /app
COPY contracts/gen/go/ contracts/gen/go/
COPY services/catalog/ services/catalog/
WORKDIR /app/services/catalog
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/catalog ./cmd/server

FROM alpine:3.23
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=build /out/catalog /usr/local/bin/catalog
EXPOSE 9001 9002
ENTRYPOINT ["catalog"]
