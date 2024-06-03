GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_TIME = $(shell TZ=Asia/Shanghai date +'%Y-%m-%d.%H:%M:%S%Z')
BUILD_FLAGS = -ldflags "-X 'computing-api/common/version.BuildFlag=$(GIT_COMMIT)+$(BUILD_TIME)'"

computing-gw-rpc:
	go build $(BUILD_FLAGS) -o ./bin/computing-gw-rpc ./computing/cmd/rpc

computing-gw:
	go build ${BUILD_FLAGS} -o ./bin/computing-gw ./computing/cmd/

user-example:
	go build ${BUILD_FLAGS} -o ./bin/user-example ./user/backend/example

clean:
	rm -rf computing-gw-rpc computing-gw user-example

.PHONY: clean