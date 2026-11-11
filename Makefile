# Go checks for the module github.com/spartan-native/host-wails.
# The npm package has its own scripts in npm/package.json.

.PHONY: test vet live tidy

test: ## Run the Go tests (with the race detector)
	go test -race ./...

vet: ## Vet the darwin, windows and linux code from any OS
	go vet ./...
	GOOS=windows go vet ./...
	GOOS=linux go vet -tags purego ./...

live: ## Print what this machine's OS reports
	SN_LIVE=1 go test -run TestLiveRead -v .

tidy:
	go mod tidy
