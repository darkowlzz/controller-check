test:
	go test -v -coverprofile cover.out ./... -count 1

tidy:
	go mod tidy -compat=1.17 -v
