.phony: upgrade-install
upgrade-install:
	helm upgrade --install \
		feature-reaper \
		./charts/feature-reaper \
		--namespace=feature-reaper \
		--create-namespace \
		--set image=foo5 

.phony: build-docker
build-docker:
	docker build -t feature-reaper:latest .

.phony: load
load: build-docker
	kind load docker-image feature-reaper:latest

