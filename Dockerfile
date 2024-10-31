# Use official Golang image as the base
FROM golang:1.23.2-alpine

# Set the working directory
WORKDIR /app

# Copy the Go modules and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Expose the port
EXPOSE 8080

# Command to run the application
CMD ["./main"]
