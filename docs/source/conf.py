# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
# import os
# import sys
# sys.path.insert(0, os.path.abspath('.'))


# -- Project information -----------------------------------------------------

project = 'Seldon Core V2'
copyright = '2021-2022 Seldon Technologies Ltd.'
author = 'smv'


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    'myst_parser',
    'sphinx_panels',
    'sphinx_tabs.tabs',
    'notfound.extension',
    'sphinx_toolbox.collapse',
    'sphinx_copybutton',
    'sphinx_togglebutton',
]

sphinx_tabs_disable_tab_closing = True
myst_enable_extensions = [
    "amsmath",
    "colon_fence",
    "deflist",
    "dollarmath",
    "html_admonition",
    "html_image",
    "linkify",
    "replacements",
    "smartquotes",
    "substitution",
    "tasklist",
]

myst_substitutions = {
    'seldon-deploy-version': '1.4.1',
    'seldon-core-artifact-version': '1.13.1',
}


source_suffix = ['.rst', '.md']
# Add any paths that contain templates here, relative to this directory.
templates_path = ['_templates']

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = []

notfound_context = {
    'title': 'Not found',
    'body': '<h1>Sorry. Seldon Core V2 page was not found!</h1><br><h4>Please use the search bar above to find what you were looking for.</h4>',
}
# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = 'sphinx_material'

# Set link name generated in the top bar.
html_title = 'Seldon Core V2'

# Material theme options (see theme.conf for more information)
html_theme_options = {

    # Set the name of the project to appear in the navigation.
    # 'nav_title': '',

    # Set you GA account ID to enable tracking
    'google_analytics_account': 'UA-54780881-6',

    # Specify a base_url used to generate sitemap.xml. If not
    # specified, then no sitemap will be built.
    # 'base_url': 'https://project.github.io/project',

    # Set the color and the accent color
    'color_primary': 'indigo',
    'color_accent': 'teal',

    # Set the repo location to get a badge with stats
    # 'repo_url': 'https://github.com/project/project/',
    # 'repo_name': 'Project',

    # Visible levels of the global TOC; -1 means unlimited
    'globaltoc_depth': 4,
    # If False, expand all TOC entries
    'globaltoc_collapse': True,
    # If True, show hidden TOC entries
    'globaltoc_includehidden': False,

    # "logo_icon": "&#xe869",

    "nav_links": [
        {"href": "contents/getting-started/index",
            "internal": True, "title": "Getting Started"},
    ],

    "heroes": {
        "index": "Documentation master"
    },

    "version_dropdown": True,
    "version_json": "_static/versions.json"
}

html_logo = '_static/images/logo.svg'
html_favicon = '_static/images/favicon.ico'
html_show_sourcelink = False
html_sidebars = {
    "**": ["globaltoc.html", "localtoc.html", "searchbox.html"],
    "index": ["searchbox.html"]
}
# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ['_static']

# Hack to make substitutions work in code-blocks https://github.com/sphinx-doc/sphinx/issues/4054#issuecomment-329097229


def ultimateReplace(app, docname, source):
    result = source[0]
    for key in app.config.ultimate_replacements:
        result = result.replace(key, app.config.ultimate_replacements[key])
    source[0] = result


ultimate_replacements = {
    "{{ seldon-deploy-version }}": myst_substitutions["seldon-deploy-version"],
    "{{ seldon-core-artifact-version }}": myst_substitutions["seldon-core-artifact-version"],
}


def setup(app):
    app.add_config_value('ultimate_replacements', {}, True)
    app.connect('source-read', ultimateReplace)
