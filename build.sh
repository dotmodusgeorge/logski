echo "--> Building"
mkdir -p bin
go build -o ./bin ./...
echo "--> Created Binaries:"
ls ./bin/