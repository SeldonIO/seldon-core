mkdir tmp
cd tmp
kustomize build ../../config/spartakus > tt.yaml
csplit --suppress-matched   tt.yaml "/^---/" "{*}"
python ../split_resources.py --folder ../../../helm-charts/seldon-core-operator/templates
