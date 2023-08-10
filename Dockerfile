# Use an official Go runtime as a parent image
FROM golang:1.20.6-alpine3.17

RUN mkdir /app

ADD . /app

# Set the working directory inside the container
WORKDIR /app

# Build the Go application
RUN go build -o main .

# Command to run the executable
CMD ["/app/main"]
