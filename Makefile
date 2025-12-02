run:
	go run cmd/api/main.go

sqlc:
	sqlc generate

# migrate-up:
# 	# Add migration command here if using a tool like migrate
# 	echo "Migrations disabled by user request"

test:
	go test ./...
