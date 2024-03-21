# === Build Stage ===
FROM golang:1.21.5 AS build

# Set the working directory
WORKDIR /app

# Define build arguments for ldflags

# Copy the Go module files and download dependencies
COPY go.mod go.sum /app/
RUN go mod download

# Copy the contents of the current directory into the build container
COPY . /app

RUN go build -o keploy .

# Build the keploy binary
# === Runtime Stage ===
FROM debian:bookworm-slim

# Update the package lists and install required packages
RUN apt-get update
RUN apt-get install -y ca-certificates curl sudo && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Docker engine
RUN curl -fsSL https://get.docker.com -o get-docker.sh && \
    sh get-docker.sh && \
    rm get-docker.sh

# Install docker-compose to PATH
RUN apt install docker-compose -y

# Copy the keploy binary and the entrypoint script from the build container
COPY --from=build /app/keploy /app/keploy
COPY --from=build /app/entrypoint.sh /app/entrypoint.sh
COPY --from=build /app/docker-compose.yml /app/docker-compose.yml

# Make the entrypoint.sh file executable
RUN chmod +x /app/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh", "/app/keploy"]