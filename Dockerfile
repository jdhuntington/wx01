FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /wx01 ./cmd/wx01/

FROM alpine:3.21
RUN apk add --no-cache tzdata
COPY --from=build /wx01 /usr/local/bin/wx01
ENTRYPOINT ["wx01"]
