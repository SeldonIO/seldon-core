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

TAG=master

cluster-manager-dependencies: 
	cd ../cluster-manager && mvn clean project-info-reports:dependencies
	cd ../cluster-manager/target/site && \
		aws s3 cp dependencies.html s3://seldon-core-docs/${TAG}/cluster-manager/dependencies/dependencies.html --acl public-read && \
		aws s3 cp css s3://seldon-core-docs/${TAG}/cluster-manager/dependencies/css/ --acl public-read --recursive && \
		aws s3 cp images s3://seldon-core-docs/${TAG}/cluster-manager/dependencies/images/ --acl public-read --recursive

.PHONY: cluster-manager-jacoco
cluster-manager-jacoco: 
	cd ../cluster-manager && mvn clean verify
	cd ../cluster-manager/target/site && \
		aws s3 cp jacoco  s3://seldon-core-docs/${TAG}/cluster-manager/jacoco/ --acl public-read --recursive

#html:: source/_static/cluster-manager/dependencies source/_static/cluster-manager/jacoco

.PHONY: clean
clean:
	rm -rf source/_static/cluster-manager/dependencies
	rm -rf source/_static/cluster-manager/jacoco
	rm -rf source/python/api
	rm -rf _build

livehtml:
	sphinx-autobuild -b html --ignore "*~" $(ALLSPHINXOPTS) $(SOURCEDIR) $(BUILDDIR)/html

livehtml-fast:
	FAST_BUILD=true sphinx-autobuild -b html --ignore "*~" $(ALLSPHINXOPTS) $(SOURCEDIR) $(BUILDDIR)/html

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)
