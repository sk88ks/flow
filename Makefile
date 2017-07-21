test:	
	go test -v $(go list ./... | grep -v vendor) -coverprofile=coverage.txt -covermode=atomic

