EXAMPLES := $(wildcard *.ipynb)
EXAMPLES_MD := $(EXAMPLES:.ipynb=.md)

.PHONY: all
all: $(EXAMPLES_MD)

.SUFFIXES: .ipynb .md
.ipynb.md:
	jupyter nbconvert \
		$< \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output $(@F)
	./markdown-transforms.sh $(@F)

