import os
import socket
import signal
import logging

from contextlib import contextmanager
from subprocess import Popen
from tenacity import retry, wait_exponential, stop_after_attempt


class MicroserviceWrapper:
    def __init__(self, app_location, envs={}, grpc=False, tracing=False):
        self.app_location = app_location
        self.env_vars = self._env_vars(envs, grpc)
        self.cmd = self._get_cmd(tracing)

    def _env_vars(self, envs, grpc):
        env_vars = dict(os.environ)
        env_vars.update(envs)
        env_vars.update(
            {
                "PYTHONUNBUFFERED": "x",
                "PYTHONPATH": self.app_location,
                "APP_HOST": "127.0.0.1",
                "PREDICTIVE_UNIT_SERVICE_PORT": "5000",
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

        if grpc:
            env_vars["API_TYPE"] = "GRPC"

        return env_vars

    def _get_cmd(self, tracing):
        cmd = (
            "seldon-core-microservice",
            self.env_vars["MODEL_NAME"],
            self.env_vars["API_TYPE"],
            "--service-type",
            self.env_vars["SERVICE_TYPE"],
            "--persistence",
            self.env_vars["PERSISTENCE"],
        )

        if tracing:
            cmd = f"{cmd} --tracing"

        return cmd

    def __enter__(self):
        try:
            logging.info("starting: %s", " ".join(self.cmd))
            self.p = Popen(
                self.cmd, cwd=self.app_location, env=self.env_vars, preexec_fn=os.setsid
            )
            self._wait_until_ready()

            return self.p
        except Exception:
            logging.error("microservice failed to stary")
            raise RuntimeError("Server did not bind to 127.0.0.1:5000")

    @retry(wait=wait_exponential(max=10), stop=stop_after_attempt(10))
    def _wait_until_ready(self):
        s1 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        r1 = s1.connect_ex(("127.0.0.1", 5000))
        s2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        r2 = s2.connect_ex(("127.0.0.1", 6005))
        if r1 == 0 and r2 == 0:
            logging.info("microservice ready")
            return
        else:
            raise ValueError("Server not ready yet")

    def __exit__(self):
        if self.p:
            os.killpg(os.getpgid(self.p.pid), signal.SIGTERM)
