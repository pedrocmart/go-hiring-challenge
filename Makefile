tidy ::
	@go mod tidy && go mod vendor

seed ::
	@go run cmd/seed/main.go

run ::
	@go build -o server ./cmd/server && ./server

test ::
	@go test -v -count=1 -race ./... -coverprofile=coverage.out -covermode=atomic

docker-up ::
	docker compose up -d

docker-down ::
	docker compose down

run-seed ::
	@$(MAKE) run & \
	$(MAKE) seed

stop ::
	@pkill -f "./server" || true

run-seed-up ::
	@$(MAKE) docker-down && \
	$(MAKE) stop && \
	$(MAKE) docker-up && \
	sleep 5 && \
	$(MAKE) run-seed
