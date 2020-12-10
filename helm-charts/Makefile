CHARTS=seldon-core-analytics seldon-core-loadtesting seldon-single-model seldon-abtest seldon-mab seldon-od-model seldon-od-transformer seldon-core-operator


build_all:
	make clean package get_index update_index

push_all:
	gsutil cp build/index.yaml build/*.tgz gs://seldon-charts/

docs: 
	# NOTE: helm-docs will ignore the charts listed in the .helmdocsignore file
	# (in the repository root)
	helm-docs -o README.md -t ../README.md.tpl

install:
	GO111MODULE=on go get github.com/norwoodj/helm-docs/cmd/helm-docs@f66fdbd6fe

package:
	mkdir -p build && \
		helm package $(CHARTS) && \
		mv *.tgz build/

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

