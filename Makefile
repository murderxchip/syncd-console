
TARGET=bin/sd

build:
	@echo "building source..."
	@go build -o $(TARGET)
	@echo "build done"

clean:
	@rm -rf bin