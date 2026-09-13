FROM golang:1.26-alpine AS build
WORKDIR /app
COPY contracts/gen/go/ contracts/gen/go/
COPY packages/go/ packages/go/
COPY services/world-domination/ services/world-domination/
WORKDIR /app/services/world-domination
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/world-domination ./cmd/server

FROM alpine:3.23
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=build /out/world-domination /usr/local/bin/world-domination
EXPOSE 9005 9006
ENTRYPOINT ["world-domination"]
