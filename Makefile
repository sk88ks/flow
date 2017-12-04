setup:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/pierrre/gotestcover
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
 
test:
	go test -v ./...

cover:
	gotestcover -coverprofile=cover.out ./...
