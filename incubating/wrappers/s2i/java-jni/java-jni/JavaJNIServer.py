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
        # TODO: Read JAR path from somewhere
        current_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
        classpath = os.path.join(current_dir, "java", "build", "libs", "model-all.jar")

        logger.debug(f"Starting JVM with classpath {classpath}")
        # NOTE: convertStrings must be set to True to avoid an explosion of
        # interop calls when working with the returned Java strings
        jpype.startJVM(classpath=[classpath], convertStrings=True)

        # TODO: Make class name configurable
        logger.debug("Instantiating MyModel object")
        java__MyModel = jpype.JPackage("io").seldon.demo.MyModel
        self._model = java__MyModel()

    def predict_rest(self, request: bytes) -> str:
        logger.debug("Sending request to Java model")
        prediction_raw = self._model.predictREST(request)

        return prediction_raw
