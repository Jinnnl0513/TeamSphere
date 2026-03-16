.PHONY: build build-linux run migrate clean

build:
	set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=amd64&& go build -o team_sphere ./cmd/server

build-arm64:
	set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=arm64&& go build -o team_sphere_arm64 ./cmd/server

run:
	go run ./cmd/server/main.go
web:
	cd web\frontend && npm run dev
migrate:
	go run ./cmd/server --migrate

clean:
	rm -f team_sphere team_sphere.exe team_sphere_linux
