all:
	make clean package get_index update_index


package:
	mkdir -p build && \
		helm package seldon-core && \
		mv seldon-core-*.tgz build/

get_index:
	mkdir -p build && \
		wget https://storage.googleapis.com/seldon-charts/index.yaml && \
		mv index.yaml build/ && \
		cp -v ./build/index.yaml ./build/index.yaml.orig

update_index:
	cd build && \
		helm repo index --merge index.yaml --url https://storage.googleapis.com/seldon-charts ./

clean:
	rm -rfv build

