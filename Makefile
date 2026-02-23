.PHONY: test integration-tests

test:
	go test ./...

integration-tests:
	go test ./cmd -run TestIntegrationFixtures
