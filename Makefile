# This provides a convenient way to quickly run all the basic tests locally.
# When running in CI/CD (eg. Cloud Build) these steps should be separated out
# so they can run in parallel, and to log the output from each step independently.

.PHONY: default test
default:

.PHONY: filter-exports-test
test: filter-exports-test
filter-exports-test:
	$(MAKE) -C image/resources/filter-exports test

.PHONY: knfsd-agent-test
test: knfsd-agent-test
knfsd-agent-test:
	$(MAKE) -C image/resources/knfsd-agent test

.PHONY: knfsd-fsidd-test
test: knfsd-fsidd-test
knfsd-fsidd-test:
	$(MAKE) -C image/resources/knfsd-fsidd test

.PHONY: knfsd-metrics-agent-test
test: knfsd-metrics-agent-test
knfsd-metrics-agent-test:
	$(MAKE) -C image/resources/knfsd-metrics-agent test

.PHONY: netapp-exports-test
test: netapp-exports-test
netapp-exports-test:
	$(MAKE) -C image/resources/netapp-exports test

.PHONY: mig-scaler-test
test: mig-scaler-test
mig-scaler-test:
	$(MAKE) -C tools/mig-scaler test

