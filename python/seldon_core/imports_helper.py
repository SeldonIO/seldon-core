import logging
import textwrap

logger = logging.getLogger(__name__)

# Variables to check if certain extra dependencies are included or
# not
_TF_MISSING = True
_GCS_MISSING = True

try:
    import tensorflow  # noqa: F401

    _TF_MISSING = False
except ImportError:
    _TF_MISSING = True
    notice = textwrap.dedent(
        """
        Tensorflow is not installed.
        If you want to use `tftensor` and Tensorflow's data types
        install `tensorflow` or install `seldon_core` as

            $ pip install seldon_core[tensorflow]
    """
    )
    logger.info(notice)


try:
    from google.cloud import storage  # noqa: F401

    _GCS_MISSING = False
except ImportError:
    _GCS_MISSING = True
    notice = textwrap.dedent(
        """
        Support for Google Cloud Storage is not installed.
        If you want to download resources from Google Cloud Storage
        install `google-cloud-storage` or install `seldon_core` as

            $ pip install seldon_core[gcs]
    """
    )
    logger.info(notice)
