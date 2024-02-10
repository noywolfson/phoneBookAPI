# Use the official Golang image
FROM golang:1.19-alpine as builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN GOOS=linux GOARCH=amd64 go build -o phoneBook .

EXPOSE 8080

ENTRYPOINT ["./phoneBook"]
