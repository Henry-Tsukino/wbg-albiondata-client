.PHONY: run frontend fmt validate-fmt build-windows build-linux build-darwin

# The version is required for the build targets so that self-update works:
#   make build-windows VERSION=1.5.0
# (use plain 1.5.0, no leading "v" - the GitHub release tag must match)
export VERSION

run:
	scripts/run.sh

frontend:
	cd frontend && npm ci && npm run build

fmt:
	scripts/fmt.sh

validate-fmt:
	scripts/validate-fmt.sh

build-windows:
	scripts/build-windows.sh

build-linux:
	scripts/build-linux.sh

build-darwin:
	scripts/build-darwin.sh
