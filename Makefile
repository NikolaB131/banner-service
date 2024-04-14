BINARYFILE=./bin/app
SOURCEFILE=./cmd/app/main.go

build: $(SOURCEFILE)
	go build -o $(BINARYFILE) $(SOURCEFILE)

build-linux: $(SOURCEFILE)
	GOOS=linux go build -o $(BINARYFILE) $(SOURCEFILE)

run: build
	$(BINARYFILE)

clean:
	rm $(BINARYFILE)

docker-up:
	docker compose up -d --build

docker-down-volumes:
	docker compose down -v

docker-rm:
	docker compose rm -f -s

docker-rm-volumes:
	docker compose rm -f -s -v

swagger-from-task:
	docker run -d -p 5000:8080 -e SWAGGER_JSON=/foo/swagger-from-task.yaml -v `pwd`:/foo swaggerapi/swagger-ui
	@echo "\033[0;32mswagger ui\033[0m started at: http://localhost:5000"

docker-run-all-tests:
	@echo Running tests...
	-docker exec banner-service-server-1 go test ./tests/e2e/...
	@echo Tests completed

e2e-tests: docker-up docker-run-all-tests docker-down-volumes
