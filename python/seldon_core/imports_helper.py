import logging
import textwrap

logger = logging.getLogger(__name__)

# Variables to check if certain extra dependencies are included or
# not
_TF_PRESENT = False

try:
    #  Fix for https://github.com/SeldonIO/seldon-core/issues/1076
    #
    #  If we do `import tensorflow` and there is a folder on the current path
    #  also named `tensorflow`, it may give a false positive even if the folder
    #  is not a Python package. This happens because, since PEP 420 got
    #  introduced in Python 3.3, folders without `__init__.py` can still get
    #  imported as namespaces.
    #
    #  To avoid this and make the check more robust, we test the presence of
    #  the `make_ndarray` method inside the `tensorflow` import.
    from tensorflow import make_ndarray  # noqa: F401

    _TF_PRESENT = True
except ImportError:
    _TF_PRESENT = False
    notice = textwrap.dedent(
        """
        Tensorflow is not installed.
        If you want to use `tftensor` and Tensorflow's data types
        install `tensorflow` or install `seldon_core` as

            $ pip install seldon_core[tensorflow]

        or

            $ pip install seldon_core[all]
        """
    )
    logger.info(notice)
