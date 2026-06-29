.PHONY: build test clean run docker-build docker-push

# Defaults to Bazel unless overridden
BAZEL ?= bazel

# -- Build --
build:
	$(BAZEL) build //cmd/server:server //cmd/healthcheck:healthcheck

run: build
	./bazel-bin/cmd/server/server

# -- Test --
test:
	$(BAZEL) test //... --test_output=errors

# -- Docker --
docker-build:
	BAZEL_REMOTE_CACHE=$(BAZEL_REMOTE_CACHE) docker build -t weekly-rundown:latest .

docker-push:
	docker push weekly-rundown:latest

# -- Clean --
clean:
	$(BAZEL) clean --expunge

# -- Gazelle --
gazelle:
	$(BAZEL) run //:gazelle -- --mode=fix

gazelle-update-repos:
	$(BAZEL) run //:gazelle -- --mode=update
