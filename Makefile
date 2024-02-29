.PHONY: create-cert
create-cert:
	openssl genrsa -out ca.key 2048
	openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=He11Coder proxy CA"
	openssl genrsa -out cert.key 2048

.PHONY: add-cert-as-trusted
add-cert-as-trusted: ca.crt
	sudo cp ca.crt /usr/local/share/ca-certificates/
	sudo update-ca-certificates -v
	openssl verify ca.crt

.PHONY: docker-build
docker-build:
	sudo docker build --tag proxy-api -f Dockerfile.api .
	sudo docker build --tag https-proxy -f Dockerfile.proxy .

.PHONY: docker-compose-up
docker-compose-up: docker-build
	sudo docker compose up -d