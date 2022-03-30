set GOOS=windows
set GOARCH=amd64

go build -o="Oinkyparty-Client.exe" ./cmd/client
go build -o="Oinkyparty-Server.exe" ./cmd/server
