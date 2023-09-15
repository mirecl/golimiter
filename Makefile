test:
	@go test -coverpkg=./... ./... -coverprofile=coverage.out -covermode=atomic > /dev/null
	@echo "-----------------------------------------------------------------------------------------------------"
	@go tool cover -func coverage.out
	@echo "-----------------------------------------------------------------------------------------------------"

lint:
	@golangci-lint run ./...