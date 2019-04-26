import pytest
from subprocess import run,Popen
import signal
import subprocess
import os
import time


def build_java_images():
    run("cd ../../api-frontend && make -f Makefile.ci build push_image_private_repo", shell=True, check=True)
    run("cd ../../engine && make -f Makefile.ci build push_image_private_repo", shell=True, check=True)    
