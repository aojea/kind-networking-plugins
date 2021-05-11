all: build

build: baremetal multicluster multizone

.PHONY: baremetal
baremetal:
	go build -v -mod vendor -o ./bin/ ./baremetal/

.PHONY: multicluster
multicluster:
	go build -v -mod vendor -o ./bin/ ./multicluster/

.PHONY: multizone
multizone:
	go build -v -mod vendor -o ./bin/ ./multizone/

clean:
	rm -f ./bin/*