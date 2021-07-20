FROM golang:1.11.5-alpine
WORKDIR /app
RUN apk update && apk upgrade && apk add --no-cache make bash git

# First build dependencies so subsequent builds go faster due to layer caching.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build the app
COPY . .
