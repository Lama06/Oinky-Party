set GOOS=linux
set GOARCH=arm

go build -o="Oinkyparty-Client.exe" ./cmd/client
go build -o="Oinkyparty-Server.exe" ./cmd/server
