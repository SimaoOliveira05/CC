# --- Configuração ---
# Apenas o IP da Nave Mãe é variável. Default é localhost.
MS_IP ?= 127.0.0.1

# --- Build ---
build:
	@echo "🛠️  A compilar..."
	go build -o bin/mothership ./cmd/mothership
	go build -o bin/rover ./cmd/rover

# --- Run ---

# make run-mothership
run-mothership:
	@echo "🚀 Nave Mãe (IP Local / 0.0.0.0)"
	./bin/mothership

# make run-rover MS_IP=<IP_DA_NAVE_MAE>
run-rover:
	@echo "🤖 Rover a conectar a $(MS_IP)..."
	./bin/rover -ms-ip=$(MS_IP)

# make run-gc MS_IP=<IP_DA_NAVE_MAE>
run-gc:
	@echo "🌍 Ground Control a conectar a http://$(MS_IP):8080..."
	cd ../ground-control && \
	VITE_API_URL="http://$(MS_IP):8080" \
	npm run dev

clean:
	rm -rf bin/
	rm -f *.log
	rm -f *_metrics.json

# --- Test Mode ---

# make test-mothership
test-mothership:
	@echo "🧪 Nave Mãe em modo de teste..."
	./bin/mothership -test-mode

# make test-rover MS_IP=<IP_DA_NAVE_MAE>
test-rover:
	@echo "🧪 Rover em modo de teste a conectar a $(MS_IP)..."
	./bin/rover -ms-ip=$(MS_IP) -test-mode