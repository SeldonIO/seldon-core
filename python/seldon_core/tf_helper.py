# Variable to check if TF is present or not
_TF_MISSING = True

try:
    import tensorflow  # noqa: F401
    print("heeello")
    _TF_MISSING = False
except ImportError:
    _TF_MISSING = True
