DEPLOY=../../deploy
kubectl create -f ${DEPLOY}/cluster_role.yaml 
kubectl create -f ${DEPLOY}/role_binding.yaml
kubectl create -f ${DEPLOY}/namespace_role.yaml
kubectl create -f ${DEPLOY}/namespace_role_binding.yaml
kubectl create -f ${DEPLOY}/service_account.yaml
kubectl create -f ${DEPLOY}/crds/machinelearning.seldon.io_seldondeployment_crd.yaml
cat ${DEPLOY}/operator.yaml | tr '\n' '\f' | sed -e "s/        - name: WATCH_NAMESPACE\f          value: ''/        - name: WATCH_NAMESPACE\f          value: 'default'"/  | tr '\f' '\n' | kubectl create -f -

