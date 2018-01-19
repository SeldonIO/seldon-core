IMAGE_REPO?=seldonio
IMAGE_NAME?=irisclassifier
IMAGE_VERSION?=0.1
MODEL_NAME?=IrisClassifier

container_image:
	docker build \
	  --build-arg IMAGE_REPO=$(IMAGE_REPO) \
	  --build-arg IMAGE_VERSION=$(IMAGE_VERSION) \
	  --build-arg MODEL_NAME=$(MODEL_NAME) \
	  --tag $(IMAGE_REPO)/$(IMAGE_NAME):$(IMAGE_VERSION) .
