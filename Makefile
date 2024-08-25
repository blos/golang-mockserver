format:
	go fmt ./*.go

build:
	go build -o bin/ *.go

build-docker:
	docker build -t mockserver:local .

run:
	MOCK_SERVER_HOST=localhost \
	MOCK_SERVER_PORT=1080 \
	go run main.go --verbose

docker: build-docker
	docker run -p "1080:1080" mockserver:local

clean:
	go clean -cache
	rm -rf mockserver