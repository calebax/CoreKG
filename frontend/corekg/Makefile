VERSION := $(shell git describe --tags | sed 's/\(.*\)-.*/\1/')

version:
	@echo "export default '$(VERSION)'" > 'src/config/version.js'

ts:
	npx tsc --noEmit

build:
	npm run build:test

dev:
	npm run dev
