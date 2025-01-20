setup:
	go mod init instance
	go mod tidy

build:
	go build -o bin/main main.go

install:
	go mod download

start:
	go run main.go

clean:
	rm -rf bin
	go mod tidy

build-docker:
	docker build -t monitoring-instance .

deploy-docker: build-docker
	docker tag monitoring-instance babutabhavya/instance-monitoring:latest
	docker push babutabhavya/instance-monitoring:latest