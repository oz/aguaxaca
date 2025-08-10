# Build the binary
FROM golang:1.24-alpine AS build

ENV PATH=/usr/local/go/bin:$PATH \
    GOLANG_VERSION=1.24.6

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -ldflags  "-s -w" -o /app/aguaxaca

# Use a ligther image for runtime.
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=build /app/aguaxaca /usr/bin/aguaxaca

RUN mkdir /data
VOLUME ["/data"]
WORKDIR /data

ENV ANTHROPIC_API_KEY="secret" \
    NITTER_HOST="http://nitter" \
    NITTER_ACCOUNT="SOAPA_Oax"

USER nobody
CMD ["/usr/bin/aguaxaca", "-listen", ":8080", "server"]
