import logging
import textwrap

logger = logging.getLogger(__name__)

# Variables to check if certain extra dependencies are included or
# not
_TF_PRESENT = False
_GCS_PRESENT = False

try:
    import tensorflow  # noqa: F401

    _TF_PRESENT = True
except ImportError:
    _TF_PRESENT = False
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

    _GCS_PRESENT = True
except ImportError:
    _GCS_PRESENT = False
    notice = textwrap.dedent(
        """
        Support for Google Cloud Storage is not installed.
        If you want to download resources from Google Cloud Storage
        install `google-cloud-storage` or install `seldon_core` as

            $ pip install seldon_core[gcs]
    """
    )
    logger.info(notice)
