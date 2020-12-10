import os
import logging
import jpype

from seldon_core.user_model import SeldonComponent

logger = logging.getLogger(__name__)


class JavaJNIServer(SeldonComponent):
    def __init__(self):
        super().__init__()
        self._model = None

    def load(self):
        """
        We can only have a single JVM per process.
        More details can be found here:
        https://jpype.readthedocs.io/en/latest/userguide.html#multiprocessing
        """
        model_jar_path = os.getenv("JAVA_JAR_PATH")

        logger.debug(f"Starting JVM with jar: {model_jar_path}")
        # NOTE: convertStrings must be set to True to avoid an explosion of
        # interop calls when working with the returned Java strings
        jpype.startJVM(classpath=[model_jar_path], convertStrings=True)

        model_import_path = os.getenv("JAVA_IMPORT_PATH")
        logger.debug(f"Instantiating {model_import_path} object")
        java__SeldonComponent = self._import_model(model_import_path)
        self._model = java__SeldonComponent()

    def _import_model(self, model_import_path: str):
        packages = model_import_path.split(".")
        if len(packages) < 2:
            raise RuntimeError(f"Invalid Java import path: {model_import_path}")

        root = packages[0]
        current_package = jpype.JPackage(root)
        for package in packages[1:]:
            current_package = getattr(current_package, package)

        return current_package

    def predict_rest(self, request: bytes) -> bytes:
        logger.debug("Sending request to Java model")
        prediction_raw = self._model.predictRawREST(request)

        return prediction_raw
