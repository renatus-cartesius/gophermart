MOCKS_DESTINATION=mocks
.PHONY: mocks
# put the files with interfaces you'd like to mock in prerequisites
# wildcards are allowed

MOCKS_SOURCE_FILES=accrual/accrual.go loyalty/loyalty.go

mocks:
	@rm -rf internal/$(MOCKS_DESTINATION)
	@for file in $(MOCKS_SOURCE_FILES); do echo "Generating mocks for internal/$$file" ; mockgen -source=internal/$$file -destination=internal/$(MOCKS_DESTINATION)/$$file; done

.PHONY: server-run
server-run:
	@go run cmd/gophermart/main.go

.PHONY: server-build
server-build:
	@(cd cmd/gophermart && go build -buildvcs=false -o gophermart)
