VERSION := $(or ${VERSION}, $(shell git describe --tags --always --first-parent))

# ============================================================
# Docker build
# ============================================================
.PHONY: docker-build
docker-build: docker-build-transformer docker-build-api docker-build-pyfunc docker-build-batch-predictor

.PHONY: docker-build-api
docker-build-api:
	$(MAKE) -C ../../ docker-build-api

.PHONY: docker-build-transformer
docker-build-transformer:
	$(MAKE) -C ../../ docker-build-transformer

.PHONY: docker-build-pyfunc
docker-build-pyfunc:
	$(MAKE) -C ../../ docker-build-pyfunc

.PHONY: docker-build-batch-predictor
docker-build-batch-predictor:
	$(MAKE) -C ../../ docker-build-batch-predictor

# ============================================================
# K3D Utilities
# ============================================================
.PHONY: k3d-import
k3d-import: k3d-import-api k3d-import-transformer k3d-import-pyfunc 

k3d-import-api:
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin:${VERSION})
	@K3D_CLUSTER=$(or ${K3D_CLUSTER},${K3D_CLUSTER},merlin-cluster)
	@k3d image import ${IMAGE_TAG} -c ${K3D_CLUSTER}

k3d-import-transformer:
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin-transformer:${VERSION})
	@K3D_CLUSTER=$(or ${K3D_CLUSTER},${K3D_CLUSTER},merlin-cluster)
	@k3d image import ${IMAGE_TAG} -c ${K3D_CLUSTER}

k3d-import-pyfunc:
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin-pyfunc-base:${VERSION})
	@K3D_CLUSTER=$(or ${K3D_CLUSTER},${K3D_CLUSTER},merlin-cluster)
	@k3d image import ${IMAGE_TAG} -c ${K3D_CLUSTER}
