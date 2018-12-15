import subprocess
from subprocess import run,Popen

def build_python_s2i_images():
    ret = run("cd ../../wrappers/s2i/python/build_scripts && ./build_all_local.sh", shell=True, check=True)

def get_s2i_python_version():
    completedProcess = Popen("cd ../../wrappers/s2i/python && grep 'IMAGE_VERSION=' Makefile | cut -d'=' -f2", shell=True, stdout=subprocess.PIPE)
    output = completedProcess.stdout.readline()
    version = output.decode('utf-8').rstrip()
    return version

