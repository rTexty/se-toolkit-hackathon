.PHONY: up down test lint seed

up:
	docker compose up --build -d

down:
	docker compose down -v

test:
	go test ./... -v -count=1 -cover

test-integration:
	go test ./tests/ -v -count=1 -tags=integration

lint:
	golangci-lint run ./...

seed:
	@echo "Seeding test data..."
	@curl -s -X POST http://localhost:8080/dummyLogin \
		-H 'Content-Type: application/json' \
		-d '{"role":"admin"}' | jq -r '.token' > /tmp/admin_token
	@TOKEN=$$(cat /tmp/admin_token); \
	curl -s -X POST http://localhost:8080/rooms/create \
		-H "Authorization: Bearer $$TOKEN" \
		-H 'Content-Type: application/json' \
		-d '{"name":"Переговорка А","capacity":6}'; \
	echo ""
	@echo "Done."
