watcher: bin/watcher

runner: bin/runner

bin/watcher: watcher/main.go
	go build -o bin/watcher ./watcher

bin/runner: runner/main.go
	go build -o bin/runner ./runner
