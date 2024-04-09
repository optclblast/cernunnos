PROJECT_DIR = $(CURDIR)
PROJECT_BIN = ${PROJECT_DIR}/bin
TOOLS_BIN = ${PROJECT_BIN}/tools

build:
	sudo docker buildx build ./Dockerfile.tests -t cernunnos:latest

up:
	sudo docker network create --driver bridge --subnet=192.168.2.0/24 --attachable cernunnos-net;
	sudo docker buildx build . -t cernunnos:latest
	sudo docker compose up -d

	sleep 5 

	sudo docker exec -it cernunnos \
		/app/cernunnos fill-db \
		-db-host=cernunnos-db:5432 \
		-db-user=cernunnos \
		-db-password=cernunnos

start:
	sudo docker network create --driver bridge --subnet=192.168.2.0/24 --attachable cernunnos-net;
	sudo docker buildx build . -t cernunnos:latest
	sudo docker compose up -d

down:
	sudo docker compose down
	sudo docker network rm cernunnos-net

netup: 
	sudo docker network create --driver bridge --subnet=192.168.2.0/24 --attachable cernunnos-net;
dropnet:
	sudo docker network rm cernunnos-net

tools:
	@GOBIN=${TOOLS_BIN} go install github.com/google/wire/cmd/wire@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${TOOLS_BIN} v1.56.2

lint:
	${TOOLS_BIN}/golangci-lint run --config ./.golangci.yaml  ./...

filldb:
	sudo docker exec -it cernunnos \
		/app/cernunnos fill-db \
		-log-level=debug \
		-db-host=cernunnos-db:5432 \
		-db-user=cernunnos \
		-db-password=cernunnos

test:
	sudo docker exec -it cernunnos 'go' 'test' '-v' '/build/tests'