build:
	@go build -o bin/ommitter .

# should implement full interactive mode first
# run: build
# 	@./bin/ommitter

tidy:
	@go mod tidy

test:
	@go test ./... -v