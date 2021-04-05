MINIKUBE_PROFILE := controller-down-test
AGONES_VERSION := 1.13.0
KUBERNETES_VERSION := 1.18.15

up:
	minikube start -p $(MINIKUBE_PROFILE) --driver virtualbox --kubernetes-version $(KUBERNETES_VERSION)
	minikube profile $(MINIKUBE_PROFILE)
	helm repo add agones https://agones.dev/chart/stable
	helm repo update
	helm upgrade --install agones --version v$(AGONES_VERSION) --namespace agones-system --create-namespace agones/agones

delete:
	minikube -p $(MINIKUBE_PROFILE) delete
