MODULE_NAME=github.com/gibsn/duty_bot

TEST_FILES = $(shell find -L * -name '*_test.go' -not -path "vendor/*")
TEST_PACKAGES = $(dir $(addprefix $(MODULE_NAME)/,$(TEST_FILES)))

all: duty_bot

duty_bot:
	go build -mod vendor -o ./bin/$@ $(MODULE_NAME)

test:
	go test -mod vendor -v $(TEST_PACKAGES)

clean:
	rm -rf ./bin
	rm -rf ./pkg

.PHONY: clean test
