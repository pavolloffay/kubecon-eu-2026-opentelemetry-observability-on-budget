REGISTRY ?= ghcr.io/pavolloffay/kubecon-eu-2026-opentelemetry-observability-on-budget
APPS = frontend backend1 backend2 backend3
KIND_CLUSTER ?= workshop
NAMESPACE ?= tutorial-application

.PHONY: docker-build docker-build-% kind-load kind-load-% deploy restart docker-build-and-load

docker-build: $(addprefix docker-build-,$(APPS))

docker-build-%:
	docker build -t $(REGISTRY)-$*:latest ./app/$*

kind-load: $(addprefix kind-load-,$(APPS))

kind-load-%:
	kind load docker-image $(REGISTRY)-$*:latest --name $(KIND_CLUSTER)

deploy:
	kubectl apply -f app/k8s.yaml

restart:
	kubectl rollout restart deployment -n $(NAMESPACE)

docker-build-and-load: docker-build kind-load restart
