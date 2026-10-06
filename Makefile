# SPDX-License-Identifier: GPL-3.0-only
ACTIONLINT_VERSION := v1.7.12
GOVULNCHECK_VERSION := v1.7.0

.PHONY: check fmt fmt-check mod-check lint test build vuln-check contributors

check: fmt-check mod-check lint test build vuln-check

fmt:
	gofmt -w cmd internal platform

fmt-check:
	@files="$$(gofmt -l cmd internal platform)" || exit 1; \
	if [ -n "$$files" ]; then printf 'Run make fmt to format:\n%s\n' "$$files"; exit 1; fi

mod-check:
	go mod tidy -diff
	go mod verify

lint:
	go vet ./...
	go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION) -shellcheck=""

test:
	go test -race -count=1 "-coverprofile=coverage.out" ./...

build:
	go build ./...

vuln-check:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

# Count authors and Co-authored-by trailers without publishing email addresses.
contributors:
	git shortlog --group=author --group=trailer:co-authored-by -sn HEAD
