# ====================================================================================
# Setup Project

PROJECT_NAME ?= provider-skycloak
PROJECT_REPO ?= github.com/decade-eng/$(PROJECT_NAME)

export TERRAFORM_VERSION ?= 1.5.7

# Do not allow a version of terraform greater than 1.5.x, due to versions 1.6+ being
# licensed under BSL, which is not permitted.
TERRAFORM_VERSION_VALID := $(shell [ "$(TERRAFORM_VERSION)" = "`printf "$(TERRAFORM_VERSION)\n1.6" | sort -V | head -n1`" ] && echo 1 || echo 0)

export TERRAFORM_PROVIDER_SOURCE ?= sky-cloak/skycloak
export TERRAFORM_PROVIDER_REPO ?= https://github.com/sky-cloak/terraform-provider-skycloak
# renovate: datasource=github-releases depName=sky-cloak/terraform-provider-skycloak
export TERRAFORM_PROVIDER_VERSION ?= 0.5.1
# Commit ref of the v$(TERRAFORM_PROVIDER_VERSION) tag; pins provider-source
# for reproducibility since tags are mutable.
export TERRAFORM_PROVIDER_COMMIT ?= v$(TERRAFORM_PROVIDER_VERSION)
export TERRAFORM_PROVIDER_DOWNLOAD_NAME ?= terraform-provider-skycloak
export TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX ?= ${TERRAFORM_PROVIDER_REPO}/releases/download/v$(TERRAFORM_PROVIDER_VERSION)
export TERRAFORM_NATIVE_PROVIDER_BINARY ?= terraform-provider-skycloak_v$(TERRAFORM_PROVIDER_VERSION)
export TERRAFORM_DOCS_PATH ?= docs/resources
export TERRAFORM_FILE_MIRROR ?= .terraform.d/plugins
export TERRAFORM_FILE_MIRROR_REPO ?= ${TERRAFORM_FILE_MIRROR}/registry.terraform.io

export GOLANGCILINT_VERSION ?= 2.13.2

PLATFORMS ?= linux_amd64 linux_arm64

# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse to
# run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

-include build/makelib/output.mk

# ====================================================================================
# Setup Go

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))

GO_REQUIRED_VERSION ?= 1.26.5
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider $(GO_PROJECT)/cmd/generator
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.Version=$(VERSION)
GO_SUBDIRS += cmd internal apis
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubernetes tools
KIND_VERSION = v0.31.0
UPTEST_VERSION = v2.2.0
CRDDIFF_VERSION = v0.12.1
CROSSPLANE_CLI_VERSION = v2.2.1
# for e2e testing
CROSSPLANE_VERSION = 2.2.1
-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Images

REGISTRY_ORGS ?= ghcr.io/decade-eng
IMAGES = $(PROJECT_NAME)
-include build/makelib/imagelight.mk

# ====================================================================================
# Setup XPKG

XPKG_REG_ORGS ?= ghcr.io/decade-eng
XPKG_REG_ORGS_NO_PROMOTE ?= ghcr.io/decade-eng
XPKGS = $(PROJECT_NAME)
-include build/makelib/xpkg.mk

# ====================================================================================
# Fallthrough

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time `make` is run, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make

# NOTE(hasheddan): we force image building to happen prior to xpkg build so that
# we ensure image is present in daemon.
xpkg.build.provider-skycloak: do.build.images

# NOTE(hasheddan): we ensure up is installed prior to running platform-specific
# build steps in parallel to avoid encountering an installation race condition.
build.init: $(UP) $(CROSSPLANE_CLI) check-terraform-version provider-source

# ====================================================================================
# Setup Terraform for fetching provider schema
TERRAFORM := $(TOOLS_HOST_DIR)/terraform-$(TERRAFORM_VERSION)
TERRAFORM_WORKDIR := $(WORK_DIR)/terraform
TERRAFORM_PROVIDER_SCHEMA := config/schema.json

check-terraform-version:
ifneq ($(TERRAFORM_VERSION_VALID),1)
	$(error invalid TERRAFORM_VERSION $(TERRAFORM_VERSION), must be less than 1.6.0 since that version introduced a not permitted BSL license))
endif

$(TERRAFORM): check-terraform-version
	@$(INFO) installing terraform $(HOSTOS)-$(HOSTARCH)
	@mkdir -p $(TOOLS_HOST_DIR)/tmp-terraform
	@curl -fsSL https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/terraform_$(TERRAFORM_VERSION)_$(SAFEHOST_PLATFORM).zip -o $(TOOLS_HOST_DIR)/tmp-terraform/terraform.zip
	@unzip $(TOOLS_HOST_DIR)/tmp-terraform/terraform.zip -d $(TOOLS_HOST_DIR)/tmp-terraform
	@mv $(TOOLS_HOST_DIR)/tmp-terraform/terraform $(TERRAFORM)
	@rm -fr $(TOOLS_HOST_DIR)/tmp-terraform
	@$(OK) installing terraform $(HOSTOS)-$(HOSTARCH)

HOST_PLATFORM := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/x86_64/amd64/;s/aarch64/arm64/')_$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

$(TERRAFORM_PROVIDER_SCHEMA): $(TERRAFORM)
	@$(INFO) generating provider schema for $(TERRAFORM_PROVIDER_SOURCE) $(TERRAFORM_PROVIDER_VERSION)
	@mkdir -p $(TERRAFORM_WORKDIR)
	@$(MAKE) download-tf-provider-platforms
	@$(MAKE) download-tf-provider-platform PLATFORM=$(HOST_PLATFORM)
	@echo '{"terraform":[{"required_providers":[{"provider":{"source":"'"$(TERRAFORM_PROVIDER_SOURCE)"'","version":"'"$(TERRAFORM_PROVIDER_VERSION)"'"}}],"required_version":"'"$(TERRAFORM_VERSION)"'"}]}' > $(TERRAFORM_WORKDIR)/main.tf.json
	@echo 'provider_installation { filesystem_mirror { path = "$(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR)" include = ["*/*/*"] } }' > $(TERRAFORM_WORKDIR)/config.tfrc
	@TF_CLI_CONFIG_FILE=$(TERRAFORM_WORKDIR)/config.tfrc $(TERRAFORM) -chdir=$(TERRAFORM_WORKDIR) init -no-color > $(TERRAFORM_WORKDIR)/terraform-logs.txt 2>&1
	@TF_CLI_CONFIG_FILE=$(TERRAFORM_WORKDIR)/config.tfrc $(TERRAFORM) -chdir=$(TERRAFORM_WORKDIR) providers schema -json=true > $(TERRAFORM_PROVIDER_SCHEMA) 2>> $(TERRAFORM_WORKDIR)/terraform-logs.txt
	@$(OK) generating provider schema for $(TERRAFORM_PROVIDER_SOURCE) $(TERRAFORM_PROVIDER_VERSION)

download-tf-provider-platforms: $(foreach p,$(PLATFORMS), download-tf-provider-platform.$(p))

download-tf-provider-platform.%:
	@$(MAKE) download-tf-provider-platform PLATFORM=$*

download-tf-provider-platform:
	@PLATFORM=$*
	@$(INFO) downloading provider for platform $(PLATFORM)
	@mkdir -p $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}
	@curl -fsSL --retry 5 --retry-delay 5 --retry-all-errors ${TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX}/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${PLATFORM}.zip -o $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/terraform.zip
	@unzip -o -qq $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/terraform.zip -d $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/
	@rm $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/terraform.zip

pull-docs:
	@if [ ! -d "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)" ]; then \
  		mkdir -p "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)" && \
		git clone -c advice.detachedHead=false --depth 1 --filter=blob:none --branch "v$(TERRAFORM_PROVIDER_VERSION)" --sparse "$(TERRAFORM_PROVIDER_REPO)" "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)"; \
	fi
	@git -C "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)" sparse-checkout set "$(TERRAFORM_DOCS_PATH)"

generate.init: $(TERRAFORM_PROVIDER_SCHEMA) pull-docs provider-source scrape-docs

.PHONY: $(TERRAFORM_PROVIDER_SCHEMA) pull-docs check-terraform-version
# ====================================================================================
# Provider source extraction for the no-fork (in-process) runtime.
#
# go.mod has a replace directive pointing at this directory so the Crossplane
# provider can construct the Terraform provider in-process. The upstream
# provider only exposes its constructor from internal/provider, so after
# cloning we inject hack/xpshim.go.tmpl as an xpshim package *inside* the
# module, which makes the internal import legal.
# third_party/ is gitignored; fresh checkouts must run `make provider-source`
# (wired into build.init, generate.init and go.modules.download) before the
# Go tree compiles.
PROVIDER_SOURCE_DIR := third_party/terraform-provider-skycloak

provider-source:
	@if [ ! -f "$(PROVIDER_SOURCE_DIR)/xpshim/shim.go" ]; then \
		$(INFO) extracting Skycloak provider source $(TERRAFORM_PROVIDER_COMMIT); \
		rm -rf "$(PROVIDER_SOURCE_DIR)"; \
		git clone -c advice.detachedHead=false --depth 1 --branch "$(TERRAFORM_PROVIDER_COMMIT)" "$(TERRAFORM_PROVIDER_REPO)" "$(PROVIDER_SOURCE_DIR)" || $(FAIL); \
		rm -rf "$(PROVIDER_SOURCE_DIR)/.git"; \
		mkdir -p "$(PROVIDER_SOURCE_DIR)/xpshim"; \
		cp hack/xpshim.go.tmpl "$(PROVIDER_SOURCE_DIR)/xpshim/shim.go"; \
		$(OK) extracting Skycloak provider source $(TERRAFORM_PROVIDER_COMMIT); \
	fi

go.modules.download: provider-source

clean: provider-source-clean
provider-source-clean:
	@rm -rf "$(PROVIDER_SOURCE_DIR)"

# Scrape resource documentation from the Terraform provider docs into
# provider-metadata.yaml, used by the generator for descriptions, examples
# and import statements.
scrape-docs: pull-docs
	@$(INFO) scraping provider metadata
	@go run github.com/crossplane/upjet/v2/cmd/scraper -n $(TERRAFORM_PROVIDER_SOURCE) -r "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_DOCS_PATH)" -o config/provider-metadata.yaml
	@$(OK) scraping provider metadata

.PHONY: provider-source provider-source-clean scrape-docs
# ====================================================================================
# End to End Testing
CROSSPLANE_NAMESPACE = crossplane-system
-include build/makelib/local.xpkg.mk
-include build/makelib/controlplane.mk

# This target requires the following environment variables to be set:
# - UPTEST_EXAMPLE_LIST, a comma-separated list of examples to test
#   To ensure the proper functioning of the end-to-end test resource pre-deletion hook, it is crucial to arrange your resources appropriately.
#   You can check the basic implementation here: https://github.com/crossplane/uptest/blob/main/internal/templates/03-delete.yaml.tmpl.
# - UPTEST_CLOUD_CREDENTIALS (optional), multiple sets of AWS IAM User credentials specified as key=value pairs.
#   The support keys are currently `DEFAULT` and `PEER`. So, an example for the value of this env. variable is:
#   DEFAULT='[default]
#   aws_access_key_id = REDACTED
#   aws_secret_access_key = REDACTED'
#   PEER='[default]
#   aws_access_key_id = REDACTED
#   aws_secret_access_key = REDACTED'
#   The associated `ProviderConfig`s will be named as `default` and `peer`.
# - UPTEST_DATASOURCE_PATH (optional), please see https://github.com/crossplane/uptest#injecting-dynamic-values-and-datasource
uptest: $(UPTEST) $(KUBECTL) $(CHAINSAW) $(CROSSPLANE_CLI)
	@$(INFO) running automated tests
	@KUBECTL=$(KUBECTL) CHAINSAW=$(CHAINSAW) CROSSPLANE_CLI=$(CROSSPLANE_CLI) CROSSPLANE_NAMESPACE=$(CROSSPLANE_NAMESPACE) $(UPTEST) e2e "${UPTEST_EXAMPLE_LIST}" --data-source="${UPTEST_DATASOURCE_PATH}" --setup-script=cluster/test/setup.sh --default-conditions="Test" || $(FAIL)
	@$(OK) running automated tests

local-deploy: build controlplane.up local.xpkg.deploy.provider.$(PROJECT_NAME)
	@$(INFO) running locally built provider
	@$(KUBECTL) wait provider.pkg $(PROJECT_NAME) --for condition=Healthy --timeout 5m
	@$(KUBECTL) -n crossplane-system wait --for=condition=Available deployment --all --timeout=5m
	@$(OK) running locally built provider

e2e: local-deploy uptest

crddiff: $(UPTEST)
	@$(INFO) Checking breaking CRD schema changes
	@for crd in $${MODIFIED_CRD_LIST}; do \
		if ! git cat-file -e "$${GITHUB_BASE_REF}:$${crd}" 2>/dev/null; then \
			echo "CRD $${crd} does not exist in the $${GITHUB_BASE_REF} branch. Skipping..." ; \
			continue ; \
		fi ; \
		echo "Checking $${crd} for breaking API changes..." ; \
		changes_detected=$$(go run github.com/crossplane/uptest/cmd/crddiff@$(CRDDIFF_VERSION) revision --enable-upjet-extensions <(git cat-file -p "$${GITHUB_BASE_REF}:$${crd}") "$${crd}" 2>&1) ; \
		if [[ $$? != 0 ]] ; then \
			printf "\033[31m"; echo "Breaking change detected!"; printf "\033[0m" ; \
			echo "$${changes_detected}" ; \
			echo ; \
		fi ; \
	done
	@$(OK) Checking breaking CRD schema changes

schema-version-diff:
	@$(INFO) Checking for native state schema version changes
	@export PREV_PROVIDER_VERSION=$$(git cat-file -p "${GITHUB_BASE_REF}:Makefile" | sed -nr 's/^export[[:space:]]*TERRAFORM_PROVIDER_VERSION[[:space:]]*:=[[:space:]]*(.+)/\1/p'); \
	echo Detected previous Terraform provider version: $${PREV_PROVIDER_VERSION}; \
	echo Current Terraform provider version: $${TERRAFORM_PROVIDER_VERSION}; \
	mkdir -p $(WORK_DIR); \
	git cat-file -p "$${GITHUB_BASE_REF}:config/schema.json" > "$(WORK_DIR)/schema.json.$${PREV_PROVIDER_VERSION}"; \
	./scripts/version_diff.py config/generated.lst "$(WORK_DIR)/schema.json.$${PREV_PROVIDER_VERSION}" config/schema.json
	@$(OK) Checking for native state schema version changes


# ====================================================================================
# Targets

# NOTE: the build submodule currently overrides XDG_CACHE_HOME in order to
# force the Helm 3 to use the .work/helm directory. This causes Go on Linux
# machines to use that directory as the build cache as well. We should adjust
# this behavior in the build submodule because it is also causing Linux users to
# duplicate their build cache, but for now we just make it easier to identify
# its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	UPBOUND_CONTEXT="local" $(GO_OUT_DIR)/provider --debug
