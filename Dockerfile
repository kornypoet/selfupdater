FROM golang:1.24-alpine AS build

WORKDIR /build
COPY go.mod go.sum .
COPY server/ ./server
RUN go build -o app ./server

FROM golang:1.24-alpine
COPY --from=build /build/app /app

ENTRYPOINT ["/app"]
