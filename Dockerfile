FROM golang:alpine AS build

ARG APP_VERSION=dev

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go tool templ generate
RUN CGO_ENABLED=1 go build -ldflags="-w -s -X catgoose/harmony/internal/version.Version=${APP_VERSION} -X catgoose/harmony/internal/version.BuildDate=$(date -u +%Y-%m-%d)" -o /harmony .

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /harmony /usr/local/bin/harmony

ENV SERVER_LISTEN_PORT=3000
EXPOSE 3000

ENTRYPOINT ["harmony"]
