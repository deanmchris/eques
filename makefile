BINARY_NAME=eques-1.0.0

build:
	GOARCH=amd64 GOAMD64=v1 go build -o ${BINARY_NAME}-default eques/eques.go
	GOARCH=amd64 GOAMD64=v2 go build -o ${BINARY_NAME}-popcnt eques/eques.go
	GOARCH=amd64 GOAMD64=v3 go build -o ${BINARY_NAME}-avx2 eques/eques.go
	GOARCH=amd64 GOAMD64=v4 go build -o ${BINARY_NAME}-avx512 eques/eques.go

build-windows:
	set GOARCH=amd64&& set GOAMD64=v1&& go build -o ${BINARY_NAME}-default.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v2&& go build -o ${BINARY_NAME}-popcnt.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v3&& go build -o ${BINARY_NAME}-avx2.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v4&& go build -o ${BINARY_NAME}-avx512.exe eques/eques.go

build-all-default:
	GOARCH=amd64 GOOS=windows go build -o ${BINARY_NAME}-windows.exe eques/eques.go
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin eques/eques.go
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux eques/eques.go

build-all-default-windows:
	set GOARCH=amd64&& set GOOS=windows&& go build -o ${BINARY_NAME}-windows.exe eques/eques.go
	set GOARCH=amd64&& set GOOS=darwin&& go build -o ${BINARY_NAME}-darwin eques/eques.go
	set GOARCH=amd64&& set GOOS=linux&& go build -o ${BINARY_NAME}-linux eques/eques.go

build-all:
	GOARCH=amd64 GOAMD64=v1 GOOS=windows go build -o ${BINARY_NAME}-windows-default.exe eques/eques.go
	GOARCH=amd64 GOAMD64=v2 GOOS=windows go build -o ${BINARY_NAME}-windows-popcnt.exe eques/eques.go
	GOARCH=amd64 GOAMD64=v3 GOOS=windows go build -o ${BINARY_NAME}-windows-avx2.exe eques/eques.go
	GOARCH=amd64 GOAMD64=v4 GOOS=windows go build -o ${BINARY_NAME}-windows-avx512.exe eques/eques.go
	GOARCH=amd64 GOAMD64=v1 GOOS=darwin  go build -o ${BINARY_NAME}-darwin-default eques/eques.go
	GOARCH=amd64 GOAMD64=v2 GOOS=darwin  go build -o ${BINARY_NAME}-darwin-popcnt eques/eques.go
	GOARCH=amd64 GOAMD64=v3 GOOS=darwin  go build -o ${BINARY_NAME}-darwin-avx2 eques/eques.go
	GOARCH=amd64 GOAMD64=v4 GOOS=darwin  go build -o ${BINARY_NAME}-darwin-avx512 eques/eques.go
	GOARCH=amd64 GOAMD64=v1 GOOS=linux   go build -o ${BINARY_NAME}-linux-default eques/eques.go
	GOARCH=amd64 GOAMD64=v2 GOOS=linux   go build -o ${BINARY_NAME}-linux-popcnt eques/eques.go
	GOARCH=amd64 GOAMD64=v3 GOOS=linux   go build -o ${BINARY_NAME}-linux-avx2 eques/eques.go
	GOARCH=amd64 GOAMD64=v4 GOOS=linux   go build -o ${BINARY_NAME}-linux-avx512 eques/eques.go

build-all-windows:
	set GOARCH=amd64&& set GOAMD64=v1&& set GOOS=windows&& go build -o ${BINARY_NAME}-windows-default.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v2&& set GOOS=windows&& go build -o ${BINARY_NAME}-windows-popcnt.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v3&& set GOOS=windows&& go build -o ${BINARY_NAME}-windows-avx2.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v4&& set GOOS=windows&& go build -o ${BINARY_NAME}-windows-avx512.exe eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v1&& set GOOS=darwin&&  go build -o ${BINARY_NAME}-darwin-default eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v2&& set GOOS=darwin&&  go build -o ${BINARY_NAME}-darwin-popcnt eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v3&& set GOOS=darwin&&  go build -o ${BINARY_NAME}-darwin-avx2 eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v4&& set GOOS=darwin&&  go build -o ${BINARY_NAME}-darwin-avx512 eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v1&& set GOOS=linux&&   go build -o ${BINARY_NAME}-linux-default eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v2&& set GOOS=linux&&   go build -o ${BINARY_NAME}-linux-popcnt eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v3&& set GOOS=linux&&   go build -o ${BINARY_NAME}-linux-avx2 eques/eques.go
	set GOARCH=amd64&& set GOAMD64=v4&& set GOOS=linux&&   go build -o ${BINARY_NAME}-linux-avx512 eques/eques.go

clean-build:
	go clean
	rm ${BINARY_NAME}-default
	rm ${BINARY_NAME}-popcnt
	rm ${BINARY_NAME}-avx2
	rm ${BINARY_NAME}-avx512

clean-build-windows:
	go clean
	del /f ${BINARY_NAME}-default.exe
	del /f ${BINARY_NAME}-popcnt.exe
	del /f ${BINARY_NAME}-avx2.exe
	del /f ${BINARY_NAME}-avx512.exe

clean-all-default:
	go clean
	rm ${BINARY_NAME}-darwin
	rm ${BINARY_NAME}-linux
	rm ${BINARY_NAME}-windows

clean-all-default-windows:
	go clean
	del /f ${BINARY_NAME}-darwin
	del /f ${BINARY_NAME}-linux
	del /f ${BINARY_NAME}-windows.exe

clean-all:
	go clean
	rm ${BINARY_NAME}-windows-default.exe
	rm ${BINARY_NAME}-windows-popcnt.exe
	rm ${BINARY_NAME}-windows-avx2.exe
	rm ${BINARY_NAME}-windows-avx512.exe
	rm ${BINARY_NAME}-darwin-default
	rm ${BINARY_NAME}-darwin-popcnt
	rm ${BINARY_NAME}-darwin-avx2
	rm ${BINARY_NAME}-darwin-avx512
	rm ${BINARY_NAME}-linux-default
	rm ${BINARY_NAME}-linux-popcnt
	rm ${BINARY_NAME}-linux-avx2
	rm ${BINARY_NAME}-linux-avx512

clean-all-windows:
	go clean
	del /f ${BINARY_NAME}-windows-default.exe
	del /f ${BINARY_NAME}-windows-popcnt.exe
	del /f ${BINARY_NAME}-windows-avx2.exe
	del /f ${BINARY_NAME}-windows-avx512.exe
	del /f ${BINARY_NAME}-darwin-default
	del /f ${BINARY_NAME}-darwin-popcnt
	del /f ${BINARY_NAME}-darwin-avx2
	del /f ${BINARY_NAME}-darwin-avx512
	del /f ${BINARY_NAME}-linux-default
	del /f ${BINARY_NAME}-linux-popcnt
	del /f ${BINARY_NAME}-linux-avx2
	del /f ${BINARY_NAME}-linux-avx512