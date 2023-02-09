import os
from itertools import chain

from setuptools import find_packages, setup

# Extra dependencies, with special 'all' key
extras = {"tensorflow": ["tensorflow"]}
all_extra_deps = chain.from_iterable(extras.values())
extras["all"] = list(set(all_extra_deps))

setup(
    name="seldon-core",
    author="Seldon Technologies Ltd.",
    author_email="hello@seldon.io",
    version="1.15.0",
    description="Seldon Core client and microservice wrapper",
    url="https://github.com/SeldonIO/seldon-core",
    license="Apache 2.0",
    license_files=["LICENSE"],
    packages=find_packages(),
    include_package_data=True,
    python_requires=">=3.6",
    install_requires=[
        "Flask >= 2.0.0, <3.0.0",
        "jsonschema<4.0.0",
        "Flask-cors<4.0.0",
        "requests<3.0.0",
        "numpy<2.0.0",
        "flatbuffers<2.0.0",
        "protobuf<4.0.0",
        "grpcio<2.0.0",
        "Flask-OpenTracing >= 1.1.0, < 1.2.0",
        "opentracing >= 2.2.0, < 2.5.0",
        "jaeger-client >= 4.1.0, < 4.5.0",
        "grpcio-opentracing >= 1.1.4, < 1.2.0",
        "grpcio-reflection < 1.35.0",
        "gunicorn >= 19.9.0, < 20.2.0",
        "setuptools >= 41.0.0",
        "prometheus_client >= 0.7.1, < 0.9.0",
        "werkzeug >= 2.1.1, < 2.3",
        "cryptography >= 3.4, < 3.5",
        # Addresses CVE SNYK-PYTHON-PYYAML-590151
        "PyYAML >= 5.4, < 5.5",
        # Addresses CVE PRISMA-2021-0020
        "click >= 8.0.0a1, < 8.1",
        # Addresses CVE CVE-2019-11236 and CVE-2020-26137 and SNYK-PYTHON-URLLIB3-1533435
        "urllib3 >= 1.26.5, < 1.27",
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
