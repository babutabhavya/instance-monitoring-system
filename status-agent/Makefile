clean:
	go mod tidy

build:
	go build -o bin/main main.go

download:
	go mod download

build-docker:
	docker build -t status-agent .

deploy-docker: build-docker
	docker tag status-agent babutabhavya/status-sync-service:latest
	docker push babutabhavya/status-sync-service:latest
