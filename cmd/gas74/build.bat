SET APP_DIR=%GOPATH%/src/github.com/fpawel/goanalit/cmd/gas74
SET GOARCH=386
buildmingw32 go build -o %APP_DIR%\gas74.exe github.com/fpawel/goanalit/cmd/gas74
