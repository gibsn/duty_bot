MODULE_NAME=github.com/gibsn/duty_bot

all: duty_bot

duty_bot:
	go build -mod vendor -o ./bin/$@ $(MODULE_NAME)

clean:
	rm -rf ./bin
	rm -rf ./pkg

.PHONY: clean
