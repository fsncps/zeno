build:
	go build -o build/zeno ./cmd/zeno

install:
	sudo install -m755 build/zeno /usr/local/bin/zeno

