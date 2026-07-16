GO      ?= go
TF      ?= terraform
BINARY  := terraform-provider-freeipa

# Nazwa wolumenu i ścieżka do przechowywania czystego stanu
VOLUME_NAME   := beeripa_freeipa-data
SNAPSHOT_FILE := .ipa-clean-volume.tar.gz

.PHONY: build test test-unit test-acc test-all docker-up docker-down clean snapshot-create snapshot-restore

build:
	$(GO) build -o $(BINARY)

test-unit:
	$(GO) test -v -count=1 ./client/... ./provider/...

# Tworzy złoty snapshot wolumenu. Uruchamiane tylko raz przy pierwszym setupie.
$(SNAPSHOT_FILE):
	@echo "==> [Inicjalizacja] Tworzenie pierwszego, czystego środowiska FreeIPA (to potrwa dłuższą chwilę)..."
	docker compose down -v
	docker compose up -d --wait
	@echo "==> Oczekiwanie na pełną gotowość API..."
	@until docker exec freeipa-test-server ipa ping >/dev/null 2>&1; do sleep 2; done
	@echo "==> Zatrzymywanie kontenera w celu bezpiecznego wykonania migawki..."
	docker compose stop
	@echo "==> Archiwizacja wolumenu do $(SNAPSHOT_FILE)..."
	docker run --rm -v $(VOLUME_NAME):/source:ro -v $(PWD):/backup alpine tar -czf /backup/$(SNAPSHOT_FILE) -C /source .
	@echo "==> Snapshot utworzony pomyślnie!"

snapshot-create:
	rm -f $(SNAPSHOT_FILE)
	$(MAKE) $(SNAPSHOT_FILE)

# Błyskawiczne przywrócenie bazy do czystego stanu przed testami
snapshot-restore: $(SNAPSHOT_FILE)
	@echo "==> Przywracanie wolumenu do czystego stanu początkowego..."
	docker compose down
	# Szybkie czyszczenie wolumenu i wypakowanie archiwum bezpośrednio przez kontener pomocniczy
	docker run --rm -v $(VOLUME_NAME):/dest -v $(PWD):/backup alpine sh -c "rm -rf /dest/* && tar -xzf /backup/$(SNAPSHOT_FILE) -C /dest"

docker-up: snapshot-restore
	@echo "==> Uruchamianie kontenera z przywróconego stanu..."
	docker compose up -d --wait
	@echo "==> Oczekiwanie na gotowość API FreeIPA..."
	@until docker exec freeipa-test-server ipa ping >/dev/null 2>&1; do sleep 1; done
	@echo "==> FreeIPA jest gotowa do testów!"

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