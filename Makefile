# NasUM - Makefile Principal

# Configuracao
MS_IP ?= 127.0.0.1
SRC_DIR = src
GC_DIR = ground-control
BIN_DIR = $(SRC_DIR)/bin
ASSETS_DIR = assets

# --- Build ---

build:
	@echo "A compilar..."
	cd $(SRC_DIR) && go build -o bin/mothership ./cmd/mothership
	cd $(SRC_DIR) && go build -o bin/rover ./cmd/rover

# --- Run ---

run-mothership: build
	@echo "Nave Mae (IP Local / 0.0.0.0)"
	cd $(SRC_DIR) && ./bin/mothership

run-rover: build
	@echo "Rover a conectar a $(MS_IP)..."
	cd $(SRC_DIR) && ./bin/rover -ms-ip=$(MS_IP)

run-gc:
	@echo "Ground Control a conectar a http://$(MS_IP):8080..."
	cd $(GC_DIR) && VITE_API_URL="http://$(MS_IP):8080" npm run dev

# --- Test Mode ---

test-mothership: build
	@echo "Nave Mae em modo de teste..."
	cd $(SRC_DIR) && ./bin/mothership -test-mode

test-rover: build
	@echo "Rover em modo de teste a conectar a $(MS_IP)..."
	cd $(SRC_DIR) && ./bin/rover -ms-ip=$(MS_IP) -test-mode

# --- Setup ---

setup-gc:
	@echo "A instalar dependencias do Ground Control..."
	cd $(GC_DIR) && npm install

# --- Clean ---

clean:
	rm -rf $(BIN_DIR)
	rm -f logs/*.log
	rm -f metrics/*.json

# --- Help ---

help:
	@echo "Comandos disponiveis:"
	@echo "  make build           - Compila mothership e rover"
	@echo "  make run-mothership  - Inicia a nave-mae"
	@echo "  make run-rover       - Inicia um rover (MS_IP=<ip>)"
	@echo "  make run-gc          - Inicia o dashboard (MS_IP=<ip>)"
	@echo "  make test-mothership - Nave-mae em modo teste"
	@echo "  make test-rover      - Rover em modo teste"
	@echo "  make setup-gc        - Instala deps do Ground Control"
	@echo "  make clean           - Limpa binarios e logs"
