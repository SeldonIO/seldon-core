
.PHONY: create
create:
	./create_templates.sh namespaced1
	./create_templates.sh spartakus
	#will add configmap parts of values file from values-patch.yaml but first ensure they're not there already
	awk '{if (f==1) { r[$$0] } else if (! ($$0 in r)) { print $$0 } } ' f=1 values-patch.yaml f=2 ../../helm-charts/seldon-core-operator/values.yaml > ./tmp/cleanvalues.yaml
	cp ./tmp/cleanvalues.yaml ../../helm-charts/seldon-core-operator/values.yaml
	cat values-patch.yaml >> ../../helm-charts/seldon-core-operator/values.yaml

.PHONY: clean
clean:
	rm -rf tmp
