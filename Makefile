.PHONY: all
all:

.PHONY: test
test:
	./hack/test

.PHONY: bench
bench:
	TEST_TYPES=benchmark ./hack/test

.PHONY: gen
gen:
	./hack/gen

.PHONY: vendor
vendor:
	$(eval $@_TMP_OUT := $(shell mktemp -d -t buildkit-bench-output.XXXXXXXXXX))
	docker buildx bake --set "*.output=type=local,dest=$($@_TMP_OUT)" vendor
	rm -rf ./vendor
	cp -R "$($@_TMP_OUT)"/* ./
	rm -rf "$($@_TMP_OUT)"/*
