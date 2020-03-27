cd ..
go build -ldflags "-s -w"
.\_build\upx -9 g.exe