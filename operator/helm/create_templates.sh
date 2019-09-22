mkdir tmp
cd tmp
kustomize build ../../config/spartakus > tt.yaml
csplit --suppress-matched   tt.yaml "/^---/" "{*}"
# manually add secret file from cert kustomize option
cp ../../config/cert/secret.yaml xxSecret
python ../split_resources.py --folder ../../../helm-charts/seldon-core-operator/templates
