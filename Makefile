.PHONY: test

test:
	go test -v --race ./...

coverage-browse:
	go test --coverprofile=cover.out ./...
	go tool cover --html=cover.out
