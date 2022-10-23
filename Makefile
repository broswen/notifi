.PHONY: compose build publish helm-template

compose:
	docker compose up --build

build:
	docker build . -f Dockerfile -t broswen/notifi:latest
	docker build . -f Dockerfile.router -t broswen/notifi-router:latest
	docker build . -f Dockerfile.delivery -t broswen/notifi-delivery:latest
	docker build . -f Dockerfile.poller -t broswen/notifi-poller:latest

publish: build
	docker push broswen/notifi:latest
	docker push broswen/notifi-router:latest
	docker push broswen/notifi-delivery:latest
	docker push broswen/notifi-poller:latest

helm-template:
	helm template api k8s/api > k8s/api.yaml
	helm template router k8s/router > k8s/router.yaml
	helm template delivery k8s/delivery > k8s/delivery.yaml
	helm template poller k8s/poller > k8s/poller.yaml

test: helm-template
	go test ./...
	kubeconform -summary -strict ./k8s/*.yaml