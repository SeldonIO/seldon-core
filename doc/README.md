# Documentation for Seldon Core

This directory contains the sources (`.md` and `.rst` files) for the
documentation. The main index page is defined in `source/index.rst`.
The Sphinx options and plugins are found in the `source/conf.py` file.
The documentation is generated in full by calling `make html`.

## Requirements

To build the documentation, first we need to install Python requirements:

```bash
make install-dev
```

### Install pandoc

We also need `pandoc` for parsing Jupyter notebooks, the easiest way
to install this is using conda:

`conda install -c conda-forge pandoc=1.19`

## Usage

To build the docs locally, you can run:

```bash
make html
```

The resulting documentation is located in the `_build` directory with
`_build/html/index.html` marking the homepage.

### Live edit

During development, is useful to have the docs server running in the background
and watching for changes on the docs.
This can be done by running:

```bash
make livehtml-fast
```

## Sphinx extensions and plugins

We use various Sphinx extensions and plugins to build the documentation:

 * [m2r2](https://github.com/crossnox/m2r2) - to handle both `.rst` and `.md`
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

### Linking to notebooks outside of `doc/source`

To reference notebooks which sit outside of the `doc` folder, you will need to
create an `*.nblink` file which links to it.
You can check the
[`source/examples/seldon_core_setup.nblink`](source/examples/seldon_core_setup.nblink)
file as an example.

Note that some notebooks may link to other resources, like images, which
usually sit relative to their folder (e.g. in a local `images/` folder).
These files also need to be referenced, so that they get linked correctly in
the final output.
We can specify these files using the `extra-media` key in the `*.nblink` file.
You can check the
[`source/examples/graph-metadata.nblink`](source/examples/seldon_core_setup.nblink)
file as an example.

The complete workaround would look like:

1. Create a simple `nblink` file pointing to the notebook:

   ```json
   {
     "path": "../../../notebooks/example-1/our-notebook.ipynb"
   }
   ```

2. (Optional) Add any extra resources (e.g. images) that the notebook links to:

   ```json
   {
     "path": "../../../notebooks/example-1/our-notebook.ipynb",
     "extra-media": ["../../../notebooks/example-1/images"]
   }
   ```

3. Reference this file instead of the included one.

