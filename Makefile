default: start

start: clean
	@echo "\nStarting dev environment:\n"
	@go build
	@./planb

build-docker: clean
	@echo "\nBuilding Docker Image:\n"
	@docker build -t planb:latest .

clean:
	@rm -f planb

release:
	@scripts/release.sh prod

release-dev:
	@scripts/release.sh dev

test:
	@go clean -testcache
	@go test -v ./backend ./log ./reverseproxy ./router ./tls . -covermode=count -coverprofile=coverage.out
