MODULE_NAME=github.com/gibsn/duty_bot

TEST_FILES = $(shell find -L * -name '*_test.go' -not -path "vendor/*")
TEST_PACKAGES = $(dir $(addprefix $(MODULE_NAME)/,$(TEST_FILES)))

TARGET_BRANCH ?= main

all: duty_bot

duty_bot:
	go build -mod vendor -o ./bin/$@ $(MODULE_NAME)

bin/golangci-lint:
	@echo "getting golangci-lint for $$(uname -m)/$$(uname -s)"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.29.0

lint: bin/golangci-lint
	bin/golangci-lint run -v -c ./build/ci/golangci.yml --new-from-rev=$(TARGET_BRANCH)

test:
	go test -mod vendor -v $(TEST_PACKAGES)

clean:
	rm -rf ./bin
	rm -rf ./pkg

.PHONY: clean test lint
