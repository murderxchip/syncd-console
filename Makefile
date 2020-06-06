TARGET=bin/sdc

build:
	@echo "building source..."
	@go build -o $(TARGET)
	@echo "build done"

clean:
	@rm -rf bin