FROM golang:1.22.8-alpine3.20

# Install make and ffmpeg
RUN apk update && apk add make ffmpeg

WORKDIR /app

# Declare a volume for data persistence

COPY . .

RUN go mod download

RUN go mod tidy

RUN make build

EXPOSE 8000

CMD ["./bin/TRANSCODING-SERVICE"]

