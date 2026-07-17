GO      ?= go
TF      ?= terraform
BINARY  := terraform-provider-freeipa

# Volume name and path for storing clean state
VOLUME_NAME   := freeipa_freeipa-data
SNAPSHOT_FILE := .ipa-clean-volume.tar.gz

.PHONY: build test test-unit test-acc test-all docker-up docker-down clean snapshot-create snapshot-restore

build:
	$(GO) build -o $(BINARY)

test-unit:
	$(GO) test -v -count=1 ./client/... ./provider/...

# Creates a golden snapshot of the volume. Run only once during initial setup.
$(SNAPSHOT_FILE):
	@echo "==> [Initialization] Creating first, clean FreeIPA environment (this will take a while)..."
	docker compose down -v
	docker compose up -d --wait
	@echo "==> Waiting for full API readiness..."
	@until docker exec freeipa-test-server ipa ping >/dev/null 2>&1; do sleep 2; done
	@echo "==> Stopping container for safe snapshot execution..."
	docker compose stop
	@echo "==> Archiving volume to $(SNAPSHOT_FILE)..."
	docker run --rm -v $(VOLUME_NAME):/source:ro -v $(PWD):/backup alpine tar -czf /backup/$(SNAPSHOT_FILE) -C /source .
	@echo "==> Snapshot created successfully!"

snapshot-create:
	rm -f $(SNAPSHOT_FILE)
	$(MAKE) $(SNAPSHOT_FILE)

# Lightning-fast restoration of database to clean state before tests
snapshot-restore: $(SNAPSHOT_FILE)
	@echo "==> Restoring volume to clean initial state..."
	docker compose down
	# Quick volume cleanup and archive extraction directly via helper container
	docker run --rm -v $(VOLUME_NAME):/dest -v $(PWD):/backup alpine sh -c "rm -rf /dest/* && tar -xzf /backup/$(SNAPSHOT_FILE) -C /dest"

docker-up: snapshot-restore
	@echo "==> Starting container from restored state..."
	docker compose up -d --wait
	@echo "==> Waiting for FreeIPA API readiness..."
	@until docker exec freeipa-test-server ipa ping >/dev/null 2>&1; do sleep 1; done
	@echo "==> FreeIPA is ready for testing!"

docker-down:
	docker compose down

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
	rm -f $(SNAPSHOT_FILE)
