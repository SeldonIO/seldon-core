import os
import logging

from typing import Dict, Union
from gunicorn.app.base import BaseApplication

logger = logging.getLogger(__name__)


def accesslog(log_level: str) -> Union[str, None]:
    """
    Enable / disable access log in Gunicorn depending on the log level.
    """

    if log_level in ["WARNING", "ERROR", "CRITICAL"]:
        return None

    return "-"


def worker_class(single_threaded: bool) -> str:
    """
    Use gthread workers to run every request on a separate thread (unless we
    explicitly force the code on a single thread).
    """

    if single_threaded:
        return "sync"

    return "gthread"


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

    def __init__(self, app, user_object, options: Dict = None):
        self.user_object = user_object
        super().__init__(app, options)

    def load(self):
        logger.debug("LOADING APP %d", os.getpid())
        try:
            logger.debug("Calling user load method")
            self.user_object.load()
        except (NotImplementedError, AttributeError):
            logger.debug("No load method in user model")
        return self.application
