@set GOOS=windows
@set GOARCH=amd64

go build -o="Oinkyparty-Client-Windows.exe" ./cmd/client
go build -o="Oinkyparty-Server-Windows.exe" ./cmd/server
