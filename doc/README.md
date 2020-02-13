# Documentation for Seldon Core

This directory contains the sources (`.md` and `.rst` files) for the
documentation. The main index page is defined in `source/index.rst`.
The Sphinx options and plugins are found in the `source/conf.py` file.
The documentation is generated in full by calling `make html`.

## Building documentation locally
To build the documentation, first we need to install Python requirements:

`pip install -r requirements_docs.txt`

We also need `pandoc` for parsing Jupyter notebooks, the easiest way
to install this is using conda:

`conda install -c conda-forge pandoc=1.19`

We are now ready to build the docs:

`make html`

The resulting documentation is located in the
`_build` directory with `_build/html/index.html` marking the homepage.

## Sphinx extensions and plugins
We use various Sphinx extensions and plugins to build the documentation:
 * [m2r](https://github.com/miyakogi/m2r) - to handle both `.rst` and `.md`
 * [sphinx.ext.napoleon](https://www.sphinx-doc.org/en/master/usage/extensions/napoleon.html) - support extracting Numpy style doctrings for API doc generation
 * [sphinx_autodoc_typehints](https://github.com/agronholm/sphinx-autodoc-typehints) - support parsing of typehints for API doc generation
 * [sphinxcontrib.apidoc](https://github.com/sphinx-contrib/apidoc) - automatic running of [sphinx-apidoc](https://www.sphinx-doc.org/en/master/man/sphinx-apidoc.html) during the build to document API
 * [nbsphinx](https://nbsphinx.readthedocs.io) - parsing Jupyter notebooks to generate static documentation
 * [nbsphinx_link](https://nbsphinx-link.readthedocs.io) - support linking to notebooks outside of Sphinx source directory via `.nblink` files

The full list of plugins and their options can be found in `source/conf.py`.

## Tips & Tricks

### Linking to markdown outside of `doc/source`

Referencing documents outside of `doc/source` tree does not work out of the box but there
is an easy workaround:

1. Create a simple `rst` file including `include` or `mdinclude` directive, e.g. see [this](source/reference/integration_nvidia_link.rst) link referenced [here](source/reference/images.md)

        .. mdinclude:: ../../../integrations/nvidia-inference-server/README.md

2. Reference this file instead of included one
