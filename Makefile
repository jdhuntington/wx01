.PHONY: build build-linux frontend dev clean deploy

# Build frontend then Go binary for current platform
build: frontend
	go build -o wx01 ./cmd/wx01/

# Cross-compile for Linux amd64 with embedded frontend
build-linux: frontend
	GOOS=linux GOARCH=amd64 go build -o wx01-linux-amd64 ./cmd/wx01/

# Build frontend assets and copy to cmd/wx01/dist for embedding
frontend:
	cd web && npm run build
	rm -rf cmd/wx01/dist
	cp -r web/dist cmd/wx01/dist

# Dev mode: run Go server (no embedded frontend — use Vite dev server separately)
dev:
	go run ./cmd/wx01/

# Package everything needed for deployment into deploy/
deploy: build-linux
	cp wx01-linux-amd64 deploy/wx01-linux-amd64
	@echo ""
	@echo "Deploy package ready in deploy/"
	@echo "Copy the deploy/ directory to your Incus container and run:"
	@echo "  sudo ./install.sh"

clean:
	rm -rf wx01 wx01-linux-amd64 cmd/wx01/dist web/dist
