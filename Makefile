
docker-build:
	go build
	@go clean

docker-up: 
	docker-compose up --force-recreate --build
