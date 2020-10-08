# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line.
SPHINXOPTS    = -j auto -n
SPHINXBUILD   = sphinx-build
SOURCEDIR     = source
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.PHONY: help Makefile

TAG=master

install-dev:
	pip install -r requirements.txt

.PHONY: clean
clean:
	rm -rf source/_static/cluster-manager/dependencies
	rm -rf source/_static/cluster-manager/jacoco
	rm -rf source/python/api
	rm -rf _build

livehtml:
	sphinx-autobuild \
		-b html \
		--ignore "*~" $(SPHINXOPTS) "$(SOURCEDIR)" "$(BUILDDIR)"

livehtml-fast:
	FAST_BUILD=true sphinx-autobuild \
		 -b html \
		 --ignore "*~" \
		 $(SPHINXOPTS) \
		 "$(SOURCEDIR)" \
		 "$(BUILDDIR)"

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile
	@$(SPHINXBUILD) \
		-M $@ \
		"$(SOURCEDIR)" \
		"$(BUILDDIR)" \
		$(SPHINXOPTS) -W --keep-going
