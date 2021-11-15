if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters"
fi

rm -rf tmp
mkdir tmp
cd tmp
kustomize build ../../config/$1 > tt.yaml
csplit --suppress-matched   tt.yaml "/^---/" "{*}"
# manually add secret file from cert kustomize option
cp ../../config/cert/secret.yaml xxSecret
python ../split_resources.py --folder ../../../helm-charts/seldon-core-operator/templates
