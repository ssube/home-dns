GOOS ?= linux
GOARCH ?= amd64

go-build:
	go build ${BUILD_OPTS} -o bin/togo-${GOOS}-${GOARCH}

go-clean:
	go clean -x

go-test: go-build
	go test ${BUILD_OPTS} -cover ./...

git-push:
	git push github
	git push gitlab

bundle-all:
	# generate hash file
	echo "path  bytes  sha256" | tee ./bin/hashes
	find ./bin/ -name "togo-*" -type f -printf '%p  %s  ' -exec sh -c 'sha256sum $$1 | cut -d " " -f 1;' find-exec {} \; \
		| tee -a ./bin/hashes
	# compress
	tar -cvf bin/togo-all.tgz ./bin/*
	gzip ./bin/togo-darwin-* || true
	gzip ./bin/togo-linux-* || true
	# do not gzip windows binaries

bundle-clean:
	rm -v ./bin/hashes || true
	rm -v ./bin/togo-* || true
