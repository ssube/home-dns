GOOS ?= linux
GOARCH ?= amd64

go-build:
	go build ${BUILD_OPTS} -o bin/home-dns-${GOOS}-${GOARCH}

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
	find ./bin/ -name "home-dns-*" -type f -printf '%p  %s  ' -exec sh -c 'sha256sum $$1 | cut -d " " -f 1;' find-exec {} \; \
		| tee -a ./bin/hashes
	# compress
	tar -cvf bin/home-dns-all.tgz ./bin/*
	gzip ./bin/home-dns-darwin-* || true
	gzip ./bin/home-dns-linux-* || true
	# do not gzip windows binaries

bundle-clean:
	rm -v ./bin/hashes || true
	rm -v ./bin/home-dns-* || true
