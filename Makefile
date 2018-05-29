build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/hello hello/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/world world/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/db db/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/triggerBuild triggerBuild/main.go
	