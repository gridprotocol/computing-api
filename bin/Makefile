GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_TIME = $(shell TZ=Asia/Shanghai date +'%Y-%m-%d.%H:%M:%S%Z')
BUILD_FLAGS = -ldflags "-X 'github.com/gridprotocol/computing-api/common/version.BuildFlag=$(GIT_COMMIT)+$(BUILD_TIME)'"

computing-api-rpc:
	go build $(BUILD_FLAGS) -o ../bin/computing-api-rpc ../computing/app/rpc

computing-api:
	go build ${BUILD_FLAGS} -o ../bin/computing-api ../computing/app/http

user-example:
	go build ${BUILD_FLAGS} -o ../bin/user-example ../user/backend/example

clean:
	rm -rf gateway-rpc gateway user-example

.PHONY: clean