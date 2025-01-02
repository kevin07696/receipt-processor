# Stage 1: Build and test
FROM alpine:latest AS build-stage

# Install Go using apk
RUN apk add --no-cache go

# Set the working directory in the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the user package
RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint /app/main.go

# Stage 2: Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian12 AS build-release-stage

# Set the working directory inside the container
WORKDIR /

# Copy the built binary from the build stage
COPY --from=build-stage /entrypoint /entrypoint
COPY --from=build-stage /app/.env /.env

# Expose the port to the outside world
EXPOSE 8080

# Create a non-root user and switch to it
USER nonroot:nonroot

# Command to run the executable
ENTRYPOINT ["/entrypoint"]
