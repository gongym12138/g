set GOOS=linux
set GOARCH=amd64
cd ..
go build -ldflags "-s -w"