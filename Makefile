watcher: bin/watcher

runner: bin/runner

collector: bin/collector

bin/watcher: watcher/main.go
	go build -o bin/watcher ./watcher

bin/runner: runner/main.go
	go build -o bin/runner ./runner

bin/collector: collector/main.go
	go build -o bin/collector ./collector
