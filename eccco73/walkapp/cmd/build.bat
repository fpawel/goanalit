set GOARCH=386
set CGO_ENABLED=1
go build -ldflags="-H windowsgui -X 'main.buildtime=%TIME%' -X main.debug=0" -o ../build/386/eccco73.exe
go build -o debug.exe -ldflags="-X 'main.buildtime=%TIME%' -X main.debug=1" -o ../build/386/debug.exe