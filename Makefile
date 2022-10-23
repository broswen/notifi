.PHONY: compose build publish helm-template

compose:
	docker compose up --build

build:
	docker build . -f Dockerfile -t broswen/notifi:latest
	#docker build . -f Dockerfile.router -t broswen/notifi-router:latest
	#docker build . -f Dockerfile.delivery -t broswen/notifi-delivery:latest

publish: build
	docker push broswen/notifi:latest
	#docker push broswen/notifi-router:latest
	#docker push broswen/notifi-delivery:latest

helm-template:
	helm template config k8s/config > k8s/config.yaml
	helm template provisioner k8s/provisioner > k8s/provisioner.yaml

test: helm-template
	go test ./...
	kubeconform -summary -strict ./k8s/*.yaml