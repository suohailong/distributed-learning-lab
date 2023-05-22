VERSION=$(shell git describe --tags --always --dirty --dirty="-dev-2")

TIME=$(shell date +%Y.%m.%d-%Hh%Mm)
PUBLISH_PATH=""
OUTPUT=harmoniakv
# ARM_MACHINE=arm64
# MIPS_MACHINE=mips64le

all: version

build:
	go build -o $(OUTPUT)


test:
	go test ./...

image: build
	docker build -t $(PUBLISH_PATH)$(OUTPUT):${VERSION}-${TIME} .

publish: image
	docker push $(PUBLISH_PATH)$(OUTPUT):${VERSION}-${TIME}

clean:
	rm -f $(OUTPUT) && rm -rf ./harmoniakv/api/v1/*.pb.go

generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/v1/harmoniakv.proto	

version:
	@echo ${VERSION}

.PHONY: build test image clean version publish


