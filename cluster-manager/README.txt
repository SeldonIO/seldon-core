Local testing:

export ENGINE_CONTAINER_IMAGE_PULL_POLICY=IfNotPresent
export SELDON_CLUSTER_MANAGER_POD_NAMESPACE=seldon
export ENGINE_CONTAINER_IMAGE_AND_VERSION=seldonio/engine:<VERSION>

mvn spring-boot:run
