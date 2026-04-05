DOCKER_IMAGE  = acho-dev
BINARY        = acho
MOUNT         = $(shell pwd):/app
DATA_DIR      = $(HOME)/.acho
SCRIPTS       = docker/scripts

DOCKER_RUN = docker run --rm -u $(shell id -u):$(shell id -g) -e HOME=/tmp -v $(MOUNT) -w /app

.PHONY: docker shell build install install-bin test fmt fmtcheck vet check vendor clean

docker:
	docker build -t $(DOCKER_IMAGE) docker/

shell:
	$(DOCKER_RUN) -it -v $(DATA_DIR):$(DATA_DIR) $(DOCKER_IMAGE) sh

build:
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/build.sh

install-bin:
	mkdir -p $(HOME)/.local/bin
	cp bin/$(BINARY) $(HOME)/.local/bin/$(BINARY)

install: build install-bin

test:
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/test.sh

fmt:
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/fmt.sh

fmtcheck:
	@$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/fmtcheck.sh

vet:
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/vet.sh

check:
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/check.sh

vendor:
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh $(SCRIPTS)/vendor.sh

clean:
	rm -rf bin/
