MODULE_NAME=github.com/gibsn/myteam_duty_bot

all: myteam_duty_bot

myteam_duty_bot:
	go build -mod vendor -o ./bin/$@ $(MODULE_NAME)

clean:
	rm -rf ./bin
	rm -rf ./pkg

.PHONY: clean
