build:
	go build
	@go clean

up: 
	@docker-compose stop
	docker-compose up --force-recreate --build

lint:
	golint ./...

format:
	find . -iname "*.go" | xargs gofmt -w -s
