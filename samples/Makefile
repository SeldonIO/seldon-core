EXAMPLES := $(wildcard *.ipynb) $(wildcard examples/*/*.ipynb)
EXAMPLES_MD := $(EXAMPLES:.ipynb=.md)

.PHONY: all
all: $(EXAMPLES_MD)

%.md: %.ipynb
	jupyter nbconvert \
		$< \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output $(@F)
	./markdown-transforms.sh $@
