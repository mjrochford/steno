
docker-build:
	go build
	@go clean

docker-up: 
	@docker-compose stop
	docker-compose up --force-recreate --build
