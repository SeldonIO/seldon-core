
.PHONY: licenses
licenses:
	./gradlew downloadLicenses --no-daemon
	rm -rf licenses
	mv build/reports/license licenses
	chmod +x ./scripts/generate_license.sh
	./scripts/generate_license.sh licenses/dependency-license.json licenses/dependency-license.txt
	cp ../../LICENSE licenses/license.txt

.PHONY: lint
lint:
	./gradlew ktlintCheck --no-daemon --no-build-cache

.PHONY: format
format:
	./gradlew ktlintFormat --no-daemon --no-build-cache

.PHONY: build
build:
	./gradlew build --no-daemon --no-build-cache