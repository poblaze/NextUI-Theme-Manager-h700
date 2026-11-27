PAK_NAME := $(shell jq -r .name pak.json)
PAK_TYPE := $(shell jq -r .type pak.json)
PAK_FOLDER := $(shell echo $(PAK_TYPE) | cut -c1)$(shell echo $(PAK_TYPE) | tr '[:upper:]' '[:lower:]' | cut -c2-)s

# Platforms the pak ships UI helpers for; each needs resources/<platform>/minui-list and minui-presenter
PLATFORMS := tg5040 h700

clean:
	rm -rf dist src/theme-manager

build:
	cd src && env CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o theme-manager cmd/theme-manager/main.go

release: build
	mkdir -p "dist/$(PAK_NAME).pak"
	$(MAKE) bump-version
	cp -R src/theme-manager resources/launch.sh README.md LICENSE pak.json "dist/$(PAK_NAME).pak"
	for platform in $(PLATFORMS); do \
		mkdir -p "dist/$(PAK_NAME).pak/bin/$$platform" && \
		cp resources/$$platform/minui-list resources/$$platform/minui-presenter "dist/$(PAK_NAME).pak/bin/$$platform/" && \
		chmod +x "dist/$(PAK_NAME).pak/bin/$$platform/"minui-* || exit 1; \
	done
	cd "dist/$(PAK_NAME).pak" && zip -r "../$(PAK_NAME).pak.zip" "."
	zip -r "dist/$(PAK_NAME).pak.zip" pak.json
	ls -lah dist

bump-version:
ifeq ($(RELEASE_VERSION),)
	error "RELEASE_VERSION is not set"
endif
	jq '.version = "$(RELEASE_VERSION)"' pak.json > pak.json.tmp
	mv pak.json.tmp pak.json
