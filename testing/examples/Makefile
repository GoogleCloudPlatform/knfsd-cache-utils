GO_FLAGS := $(GO_FLAGS) -count=1 -timeout=1h
SINGLE_STAGE := SKIP_apply=true SKIP_check=true SKIP_destroy=true

.PHONY: default
default:

# Runs all the test stages (apply, test, destroy).
# This ensures (best effort) that infrastructure is destroyed even if the tests
# fail. Running the stages individually runs the risk that on an error the
# infrastructure is not destroyed.
.PHONY: test
test:
	@go test $(GO_FLAGS)

.PHONY: apply
apply:
	@$(SINGLE_STAGE) SKIP_apply= go test $(GO_FLAGS)

# Runs only the test stage without applying or destroy Terraform. Useful when
# developing tests locally.
.PHONY: check
check:
	@$(SINGLE_STAGE) SKIP_check= go test $(GO_FLAGS)

.PHONY: destroy
destroy:
	@$(SINGLE_STAGE) SKIP_destroy= go test $(GO_FLAGS)
