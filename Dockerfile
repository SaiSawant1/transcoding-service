FROM golang:1.23.2-bookworm

RUN apt-get update && apt-get install -y make

WORKDIR /app

# Declare a volume for data persistence

COPY . .

RUN go mod download

RUN go mod tidy

RUN make build

EXPOSE 8000

CMD ["./bin/TRANSCODING-SERVICE"]
