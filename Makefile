lint:
	golangci-lint run ./...

coverage-clear:
	rm -rf coverage.out coverage.log

coverage-total:
	go tool cover -func coverage.out | grep total | awk '{print "\nTotal coverage: ", $$3}'

coverage-test:
	go test -race ./... -coverprofile=coverage.out

cover: coverage-clear coverage-test coverage-total