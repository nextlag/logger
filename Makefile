installLinter:
	brew install golangci-lint

lint: gofumpt
	golangci-lint --color always run
gofumpt:
	gofumpt -l -w .

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: installLinter lint gofumpt coverage


