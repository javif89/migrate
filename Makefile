run : cli execute

execute:
	echo $(cmd)
	cd ./bin; ./migrate $(cmd)

cli:
	go build -o bin/migrate ./cmd