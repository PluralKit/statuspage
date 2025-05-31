all: backend frontend

.PHONY: backend
backend:
	mkdir -p build
	cd backend && go build -o status . && mv ./status ../build/

.PHONY: frontend
frontend:
	rm -rf build/srv/* && mkdir -p build build/srv
	cd frontend && npm install && npm run build && mv -f build/* ../build/srv

.PHONY: clean
clean:
	rm -rf status build/

.PHONY: run
run:
	cd build && ./status