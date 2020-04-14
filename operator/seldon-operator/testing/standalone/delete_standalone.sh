DEPLOY=../../deploy
kubectl delete -f ${DEPLOY}/cluster_role.yaml
kubectl delete -f ${DEPLOY}/role_binding.yaml
kubectl delete -f ${DEPLOY}/namespace_role.yaml
kubectl delete -f ${DEPLOY}/namespace_role_binding.yaml
kubectl delete -f ${DEPLOY}/service_account.yaml
kubectl delete -f ${DEPLOY}/crds/machinelearning.seldon.io_seldondeployment_crd.yaml
kubectl delete -f ${DEPLOY}/operator.yaml

