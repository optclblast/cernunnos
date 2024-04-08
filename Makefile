PROJECT_DIR = $(CURDIR)
PROJECT_BIN = ${PROJECT_DIR}/bin
TOOLS_BIN = ${PROJECT_BIN}/tools

build:
	sudo docker buildx build . -t cernunnos:latest

up:
	sudo docker buildx build . -t cernunnos:latest
	sudo docker compose up -d

tools:
	@GOBIN=${TOOLS_BIN} go install github.com/google/wire/cmd/wire@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b ${TOOLS_BIN} v1.56.2

lint:
	${TOOLS_BIN}/golangci-lint run --config ./.golangci.yaml  ./...

filldb:
	sudo docker exec -it lamoda-tech-task-cernunnos-1 \
		/app/cernunnos fill-db \
		-log-level=debug \
		-db-host=cernunnos-db:5432 \
		-db-user=cernunnos \
		-db-password=cernunnos
