.SUFFIXES: .ipynb .md
EXAMPLES := $(wildcard *.ipynb)
EXAMPLES_MD=$(patsubst %.ipynb, %.md, $(EXAMPLES))

.PHONY: all
all: $(EXAMPLES_MD)

.ipynb.md:
	jupyter nbconvert \
		$< \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output $(notdir $@)
	./markdown-transforms.sh $(notdir $@)

