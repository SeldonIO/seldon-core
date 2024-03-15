
.PHONY: licenses
licenses:
	./gradlew downloadLicenses --no-daemon
	rm -rf licenses
	mv build/reports/license licenses
	chmod +x ./scripts/generate_license.sh
	./scripts/generate_license.sh licenses/dependency-license.json licenses/dependency-license.txt
	cp ../../LICENSE licenses/license.txt
