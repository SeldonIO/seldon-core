import logging
import os
import signal
import socket
import time
from contextlib import contextmanager
from subprocess import Popen

from tenacity import retry, retry_if_exception_type, stop_after_attempt, wait_fixed


class UserObject:
    def predict(self, X, features_names):
        logging.info("Predict called")
        return X


class MicroserviceWrapper:
    def __init__(self, app_location, envs={}, tracing=False):
        self.app_location = app_location
        self.env_vars = self._env_vars(envs)
        self.cmd = self._get_cmd(tracing)

    def _env_vars(self, envs):
        env_vars = dict(os.environ)
        env_vars.update(envs)
        env_vars.update(
            {
                "PYTHONUNBUFFERED": "x",
                "PYTHONPATH": self.app_location,
                "APP_HOST": "127.0.0.1",
                "PREDICTIVE_UNIT_HTTP_SERVICE_PORT": "9000",
                "PREDICTIVE_UNIT_GRPC_SERVICE_PORT": "5000",
                "PREDICTIVE_UNIT_METRICS_SERVICE_PORT": "6005",
                "PREDICTIVE_UNIT_METRICS_ENDPOINT": "/metrics-endpoint",
            }
        )

        s2i_env_file = os.path.join(self.app_location, ".s2i", "environment")
        with open(s2i_env_file) as fh:
            for line in fh.readlines():
                line = line.strip()
                if line:
                    key, value = line.split("=", 1)
                    key, value = key.strip(), value.strip()
                    if key and value:
                        env_vars[key] = value

        return env_vars

    def _get_cmd(self, tracing):
        cmd = (
            "seldon-core-microservice",
            self.env_vars["MODEL_NAME"],
            "--service-type",
            self.env_vars["SERVICE_TYPE"],
            "--persistence",
            self.env_vars["PERSISTENCE"],
        )

        if tracing:
            cmd += ("--tracing",)

        return cmd

    def __enter__(self):
        try:
            logging.info(f"starting: {' '.join(self.cmd)}")
            self.p = Popen(
                self.cmd, cwd=self.app_location, env=self.env_vars, preexec_fn=os.setsid
            )

            time.sleep(1)
            self._wait_until_ready()

            return self.p
        except Exception:
            logging.error("microservice failed to start")
            raise RuntimeError("Server did not bind to 127.0.0.1:5000")

    @retry(wait=wait_fixed(4), stop=stop_after_attempt(10))
    def _wait_until_ready(self):
        logging.debug("=== trying again")
        s1 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        r1 = s1.connect_ex(("127.0.0.1", 9000))
        s2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        r2 = s2.connect_ex(("127.0.0.1", 6005))
        s3 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        r3 = s3.connect_ex(("127.0.0.1", 5000))

        if r1 != 0 or r2 != 0 or r3 != 0:
            raise EOFError("Server not ready yet")

        logging.info("microservice ready")

    def _get_return_code(self):
        self.p.poll()
        return self.p.returncode

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.p:
            group_id = os.getpgid(self.p.pid)
            # Kill the entire process groups (including subprocesses of self.p)
            os.killpg(group_id, signal.SIGKILL)
