# Start from the official Golang image to build your application
FROM golang:1.16 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# Start a new stage from scratch
FROM ubuntu:20.04

# Install Streamlink
RUN apt-get update && apt-get install -y \
    python3-pip \
    && pip3 install streamlink

# Copy the built executable from the builder stage
COPY --from=builder /app/myapp .

# Expose port 8080 for your application
EXPOSE 8080

# Command to run the executable
CMD ["./myapp"]

