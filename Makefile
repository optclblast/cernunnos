PROJECT_DIR = $(CURDIR)
PROJECT_BIN = ${PROJECT_DIR}/bin
TOOLS_BIN = ${PROJECT_BIN}/tools

up:
	sudo docker compose up -d

tools:
	@GOBIN=${TOOLS_BIN} go install github.com/google/wire/cmd/wire@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${TOOLS_BIN} v1.56.2

lint:
	${TOOLS_BIN}/golangci-lint run --config ./.golangci.yaml  ./...