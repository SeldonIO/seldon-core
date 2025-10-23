import os
from itertools import chain

from setuptools import find_packages, setup

# Extra dependencies, with special 'all' key
extras = {"tensorflow": ["tensorflow==2.20.0"]}
all_extra_deps = chain.from_iterable(extras.values())
extras["all"] = list(set(all_extra_deps))

setup(
    name="seldon-core",
    author="Seldon Technologies Ltd.",
    author_email="hello@seldon.io",
    version="1.19.0-dev",
    description="Seldon Core client and microservice wrapper",
    url="https://github.com/SeldonIO/seldon-core",
    license="Business Source License 1.1",
    license_files=["LICENSE"],
    packages=find_packages(),
    include_package_data=True,
    python_requires=">=3.12",  # distutils removed in 3.12
    install_requires=[
        "flask>=3.1.2",
        "jsonschema>=4.25.1",
        "flask-cors>=6.0.1",
        "requests>=2.32.5",
        "numpy>=2.3.4",
        "protobuf>=6.33.0",
        "grpcio>=1.76.0",
        "flask-opentracing>=2.0.0",
        "opentracing>=2.4.0",  # latest release is from Nov 2020
        "jaeger-client>=4.8.0",  # latest release is from Sep 2021
        "grpcio-opentracing>=1.1.4",  # latest release is from Apr 2019
        "grpcio-reflection>=1.76.0",
        "gunicorn>=23.0.0",
        "setuptools>=80.9.0",
        "prometheus_client>=0.23.1",
        "werkzeug>=3.1.3",
        "cryptography>=46.0.3",
        "pyyaml>=6.0.3",
        "click>=8.3.0",
        "urllib3>=2.5.0",
    ],
    extras_require=extras,
    entry_points={
        "console_scripts": [
            "seldon-core-microservice = seldon_core.microservice:main",
            "seldon-core-tester = seldon_core.microservice_tester:main",
            "seldon-core-microservice-tester = seldon_core.microservice_tester:main",
            "seldon-core-api-tester = seldon_core.api_tester:main",
            "seldon-batch-processor = seldon_core.batch_processor:run_cli",
        ]
    },
    zip_safe=False,
)
