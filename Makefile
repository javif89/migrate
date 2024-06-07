run : cli execute

execute:
	cd ./bin; ./migrate $(cmd)

cli:
	go build -o bin/migrate ./cmd

test:
	go test ./... -cover