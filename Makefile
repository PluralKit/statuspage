all: backend frontend

.PHONY: backend
backend:
	mkdir -p build
	cd backend && CGO_ENABLED=1 go build -o status -ldflags "-X main.version=$(git rev-parse HEAD)" . && mv ./status ../build/

.PHONY: frontend
frontend:
	rm -rf build/srv/* && mkdir -p build build/srv
	cd frontend && npm install && npm run build && mv -f build/* ../build/srv

.PHONY: clean
clean:
	rm -rf status build/

.PHONY: dev
dev: backend
	rm -rf build/srv/* && mkdir -p build build/srv
	cd frontend && npm install && PUBLIC_API_URL="http://localhost:8080" npm run build && mv -f build/* ../build/srv
	cd build && pluralkit__status__run_dev=true ./status