BUILD_PATH = build
SRC_PATH = .
SRC_EXT = go

SOURCES = $(shell find $(SRC_PATH)/ -name '*.$(SRC_EXT)')

all: dirs $(BUILD_PATH)/nolock $(BUILD_PATH)/redlock
.PHONY: all

$(BUILD_PATH)/nolock:
	go build -o $(BUILD_PATH)/nolock nolock.go utils.go

$(BUILD_PATH)/redlock:
	go build -o $(BUILD_PATH)/redlock redlock.go utils.go

.PHONY: dirs
dirs:
	@echo "Creating directories"
	@mkdir -p $(BUILD_PATH)

clean:
	rm -rf build
.PHONY: clean
