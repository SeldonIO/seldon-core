from setuptools import setup, Command, find_packages
import os
import sys

if sys.version_info < (3, 6):
    sys.exit(
        "\nSorry, Python < 3.6 is not supported\nIf you have Python 3.x installed use: pip3 install seldon-batch"
    )
    sys.exit("")

currentFileDirectory = os.path.dirname(__file__)
with open(os.path.join(currentFileDirectory, "README.md"), "r") as f:
    readme = f.read()

# Maintain dependencies of requirements.txt
with open("requirements.txt") as f:
    requirements = f.read().splitlines()

# Maintain dependencies of requirements.txt
with open("requirements-dev.txt") as f:
    requirements_dev = f.read().splitlines()


class CleanCommand(Command):
    """Custom clean command to tidy up the project root."""

    user_options = []

    def initialize_options(self):
        pass

    def finalize_options(self):
        pass

    def run(self):
        os.system(
            "rm -vrf ./build ./dist ./*.pyc ./*.tgz ./*.egg-info ./**/__pycache__ ./__pycache__ ./.eggs ./.cache"
        )


setup(
    name="seldon-batch-processor",
    version="0.0.1",
    description="Seldon Batch Processing utility to perform gang scheduling processing across batch workloads",
    long_description=readme,
    author="SeldonIO",
    author_email="hello@seldon.io",
    url="https://github.com/SeldonIO/seldon-core",
    classifiers=[
        "Intended Audience :: Developers",
        "Natural Language :: English",
        "Programming Language :: Python :: 3.6",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
    ],
    keywords="seldon, kubernetes, batch, processing",
    license="Apache 2.0",
    packages=find_packages(exclude=["*.tests", "*.tests.*", "tests.*", "tests"]),
    include_package_data=True,
    install_requires=requirements,
    setup_requires=["pytest-runner"],
    tests_require=requirements_dev,
    test_suite="tests",
    entry_points={
        "console_scripts": ["seldon-batch-processor = seldon_batch.cmd:run_cli",]
    },
    cmdclass={"clean": CleanCommand},
)
