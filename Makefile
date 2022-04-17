BINARY_NAME=deadshot
LAST_TAG = $(shell git describe --abbrev=0 --tags)
LAST_COMMIT = $(shell git rev-parse --short HEAD)

run: generate
	@go run -tags devel ./cmd/$(BINARY_NAME) --testnet --debug ||:

release: generate
	@echo Last release version: $(LAST_TAG);
	@read -p "New release version: " release_version; \
	git tag -a -s -m "Release $${release_version}" $${release_version}; \
	git push --tags

snapshot:
	goreleaser build --snapshot --rm-dist

build: 
	goreleaser build --rm-dist

clean:
	go clean
	rm -r ./dist

test: generate
	go test ./...

generate:
	@go generate ./...
