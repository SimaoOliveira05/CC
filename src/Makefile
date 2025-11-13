build:
	go build -o bin/mothership ./cmd/mothership
	go build -o bin/rover ./cmd/rover

run-mothership: build
	@if [ -z "$(IP)" ]; then IP=127.0.0.1; fi; \
	gnome-terminal -- bash -c "./bin/mothership --mother-ip=$$IP; exec bash"

run-rover: build
	@if [ -z "$(MIP)" ]; then MIP=127.0.0.1; fi; \
	if [ -z "$(RIP)" ]; then RIP=127.0.0.1; fi; \
	gnome-terminal -- bash -c "./bin/rover --mother-ip=$$MIP --rover-ip=$$RIP; exec bash"

run-all: build
	@if [ -z "$(IPS)" ]; then IPS="127.0.0.1"; fi; \
	gnome-terminal -- bash -c "./bin/mothership --mother-ip=$$(echo $(IPS) | awk '{print $$1}'); exec bash"; \
	for ip in $(IPS); do \
		gnome-terminal -- bash -c "./bin/rover --mother-ip=$$(echo $(IPS) | awk '{print $$1}') --rover-ip=$$ip; exec bash"; \
		sleep 3; \
	done

clean:
	rm -f bin/mothership bin/rover