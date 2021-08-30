# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line.
SPHINXOPTS    = -n
SPHINXBUILD   = sphinx-build
SOURCEDIR     = source
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.PHONY: clean livehtml livehtml-fast help Makefile

TAG=master

install-dev:
	pip install -r requirements.txt
	# Install local version of seldon_core as well
	make -C ../python install_dev

clean:
	rm -rf source/_static/cluster-manager/dependencies
	rm -rf source/_static/cluster-manager/jacoco
	rm -rf source/python/api
	rm -rf _build

lint: html linkcheck

linkcheck:
	# Check for links reported as broken, ignoring 429 (too many requests)
	- LINKCHECK=true $(SPHINXBUILD) \
		-b linkcheck \
		-j auto $(SPHINXOPTS) \
		"$(SOURCEDIR)" \
		"$(BUILDDIR)"
	! grep -v '429' _build/output.txt | grep 'broken'


livehtml:
	sphinx-autobuild \
		-b html \
		--ignore "*~" -j auto $(SPHINXOPTS) "$(SOURCEDIR)" "$(BUILDDIR)"

livehtml-fast:
	FAST_BUILD=true sphinx-autobuild \
		 -b html \
		 --ignore "*~" \
		 -j auto $(SPHINXOPTS) \
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
