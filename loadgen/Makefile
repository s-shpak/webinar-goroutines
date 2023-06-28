.PHONY: build-datagen
build-datagen:
	go build -o ./cmd/datagen/datagen ./cmd/datagen

.PHONY: build-loadgen
build-loadgen:
	go build -o ./cmd/loadgen/loadgen ./cmd/loadgen

.PHONY: race-build-loadgen
race-build-loadgen:
	go build -race -o ./cmd/loadgen/loadgen ./cmd/loadgen
