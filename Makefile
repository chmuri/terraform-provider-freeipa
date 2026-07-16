GO      := /home/chmuri/.gvm/gos/go1.26.5/bin/go
TF      := /home/chmuri/.tfenv/bin/terraform
BINARY  := terraform-provider-freeipa

.PHONY: build test test-unit test-acc test-all docker-up docker-down clean

build:
	$(GO) build -o $(BINARY)

test-unit:
	$(GO) test -v -count=1 ./client/... ./provider/...

docker-up:
	docker compose up -d --wait

docker-down:
	docker compose down -v

test-acc: docker-up build
	TF_ACC=1 \
	FREEIPA_HOST=ipa.test.local \
	FREEIPA_USERNAME=admin \
	FREEIPA_PASSWORD=SecretAdminPassword123! \
	FREEIPA_INSECURE=true \
	$(GO) test -v -timeout 30m -count=1 ./provider/... -run TestAcc

test-all: test-unit test-acc

clean:
	rm -f $(BINARY)
	docker compose down -v
