import importlib

# Variable to check if TF is present or not
_TF_MISSING = importlib.util.find_spec("tensorflow") is None
