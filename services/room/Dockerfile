FROM golang:1.26-alpine AS build
WORKDIR /app
COPY contracts/gen/go/ contracts/gen/go/
COPY packages/go/ packages/go/
COPY services/room/ services/room/
WORKDIR /app/services/room
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/room ./cmd/server

FROM alpine:3.23
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=build /out/room /usr/local/bin/room
EXPOSE 9003 9004
ENTRYPOINT ["room"]
