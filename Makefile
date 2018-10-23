.PHONY: test

test:
	go test --race ./...

coverage-browse:
	go test --coverprofile=cover.out ./...
	go tool cover --html=cover.out
