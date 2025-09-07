# Use golang build image
FROM golang:1.13 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies using go modules.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
# -mod=readonly ensures immutable go.mod and go.sum in container builds.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server ./cmd/helloworld

# Use alpine as production image
FROM alpine:3
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

# Copy the binary to the production image from build image
COPY --from=builder /app/server /server

# Copy static content to container image.
COPY web/static/ /static/

# Run web service on startup.
USER app:app
CMD ["/server"]
