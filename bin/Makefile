LIB_NAME=libpow.so
GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_TIME = $(shell TZ=Asia/Shanghai date +'%Y-%m-%d.%H:%M:%S%Z')
BUILD_FLAGS = -ldflags "-X 'github.com/gridprotocol/computing-api/common/version.BuildFlag=$(GIT_COMMIT)+$(BUILD_TIME)'"

all: clean cuda computing-api

clean:
	rm -f computing-api ${LIB_NAME}

cuda:
	nvcc --ptxas-options=-v --compiler-options '-fPIC' -o ${LIB_NAME} --shared ../prover/CudaSha256/pow.cu

computing-api:
	go build ${BUILD_FLAGS} -o computing-api ../computing/app/http

computing-api-rpc:
	go build $(BUILD_FLAGS) -o computing-api-rpc ../computing/app/rpc

user-example:
	go build ${BUILD_FLAGS} -o ../bin/user-example ../user/backend/example

.PHONY: clean