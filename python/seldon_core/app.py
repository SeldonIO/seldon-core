import os
import logging
import atexit

from multiprocessing.util import _exit_function
from typing import Dict, Union
from gunicorn.app.base import BaseApplication
from seldon_core.utils import setup_tracing

logger = logging.getLogger(__name__)


def post_worker_init(worker):
    # Remove the atexit handler set up by the parent process
    # https://github.com/benoitc/gunicorn/issues/1391#issuecomment-467010209
    atexit.unregister(_exit_function)


def accesslog(log_level: str) -> Union[str, None]:
    """
    Enable / disable access log in Gunicorn depending on the log level.
    """

    if log_level in ["WARNING", "ERROR", "CRITICAL"]:
        return None

    return "-"


def threads(threads: int, single_threaded: bool) -> int:
    """
    Number of threads to run in each Gunicorn worker.
    """

    if single_threaded:
        return 1

    return threads


class StandaloneApplication(BaseApplication):
    """
    Standalone Application to run a Flask app in Gunicorn.
    """

    def __init__(self, app, options: Dict = None):
        self.application = app
        self.options = options
        super().__init__()

    def load_config(self):
        config = dict(
            [
                (key, value)
                for key, value in self.options.items()
                if key in self.cfg.settings and value is not None
            ]
        )
        for key, value in config.items():
            self.cfg.set(key.lower(), value)

    def load(self):
        return self.application


class UserModelApplication(StandaloneApplication):
    """
    Gunicorn application to run a Flask app in Gunicorn loading first the
    user's model.
    """

    def __init__(
        self, app, user_object, jaeger_extra_tags, interface_name, options: Dict = None
    ):
        self.user_object = user_object
        self.jaeger_extra_tags = jaeger_extra_tags
        self.interface_name = interface_name
        super().__init__(app, options)

    def load(self):
        if self.jaeger_extra_tags is not None:
            logger.info("Tracing branch is active")
            from flask_opentracing import FlaskTracing

            tracer = setup_tracing(self.interface_name)

            logger.info("Set JAEGER_EXTRA_TAGS %s", self.jaeger_extra_tags)
            FlaskTracing(tracer, True, self.application, self.jaeger_extra_tags)
        logger.debug("LOADING APP %d", os.getpid())
        try:
            logger.debug("Calling user load method")
            self.user_object.load()
        except (NotImplementedError, AttributeError):
            logger.debug("No load method in user model")
        return self.application
