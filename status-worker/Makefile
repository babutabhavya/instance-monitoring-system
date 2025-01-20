clean:
	go mod tidy

build:
	go build -o bin/main main.go

download:
	go mod download

build-docker:
	docker build -t status-worker .

deploy-docker: build-docker
	docker tag status-worker babutabhavya/status-request-service:latest
	docker push babutabhavya/status-request-service:latest
