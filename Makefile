build:
	@go build -buildvcs=false -o bin/TRANSCODING-SERVICE

run: build
	@./bin/TRANSCODING-SERVICE
test:
	@go test -v ./...

docker-rmi:
	@docker rmi queue:latest

docker:
	@docker run -v .:/app -p 8000:8000 transcoding:latest

docker-build:
	@docker build -t transcoding:latest .
