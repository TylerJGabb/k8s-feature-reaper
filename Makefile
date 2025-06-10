IMG=feature-reaper
KIND_CLUSTER_NAME=feature-reaper

.phony: upgrade-install
upgrade-install:
	helm upgrade --install \
		feature-reaper \
		./charts/feature-reaper \
		--namespace=feature-reaper \
		--create-namespace \
		--set image=$(IMG)

.phony: build-docker
build-docker:
	docker build -t $(IMG) .

.phony: kind-create
kind-create:
	kind create cluster --name $(KIND_CLUSTER_NAME)

.phony: kind-delete
kind-delete:
	kind delete cluster --name $(KIND_CLUSTER_NAME)

.phony: load
load: build-docker
	kind load docker-image $(IMG) --name $(KIND_CLUSTER_NAME)

.phony: unit-test
unit-test:
	go clean -testcache
	go test -v ./... -coverprofile=coverage.out -covermode=atomic

.phony: integration-test
integration-test:
	./integration-test.sh

