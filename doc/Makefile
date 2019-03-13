# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line.
SPHINXOPTS    =
SPHINXBUILD   = sphinx-build
SOURCEDIR     = source
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.PHONY: help Makefile

source/_static/cluster-manager:
	mkdir -p source/_static/cluster-manager

source/_static/cluster-manager/dependencies: source/_static/cluster-manager
	cd ../cluster-manager && mvn clean project-info-reports:dependencies
	mkdir -p source/_static/cluster-manager/dependencies
	cp -R ../cluster-manager/target/site/* source/_static/cluster-manager/dependencies

source/_static/cluster-manager/jacoco: source/_static/cluster-manager
	cd ../cluster-manager && mvn clean verify
	cp -R ../cluster-manager/target/site/jacoco source/_static/cluster-manager

#html:: source/_static/cluster-manager/dependencies source/_static/cluster-manager/jacoco

.PHONY: clean
clean:
	rm -r source/_static/cluster-manager/dependencies
	rm -r source/_static/cluster-manager/jacoco

livehtml:
	sphinx-autobuild -b html --ignore "*~" $(ALLSPHINXOPTS) $(SOURCEDIR) $(BUILDDIR)/html

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)
