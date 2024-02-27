# TODO: Do this in build.rs instead

go_possum_h := go/cpossum/possum.h

all: $(go_possum_h) README.md

.PHONY: $(go_possum_h)
$(go_possum_h):
	cbindgen -q --output $@

README.md: README.adoc
	npx downdoc $<
