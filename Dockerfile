# Build: vet + tests are gates — a failing one fails the image build.
FROM golang:1.25-alpine AS build
ENV CGO_ENABLED=0 GOFLAGS=-buildvcs=false
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go vet ./... \
 && go test ./... \
 && go build -trimpath -ldflags="-s -w" -o /out/movies .

# Runtime: ffprobe for stream details (codecs, languages, duration).
FROM alpine:3.24.1 AS runtime
RUN apk add --no-cache ffmpeg tzdata ca-certificates
COPY --from=build /out/movies /usr/local/bin/movies
ENV DATA_DIR=/config PORT=8096
EXPOSE 8096
USER 1000:1000
ENTRYPOINT ["/usr/local/bin/movies"]
