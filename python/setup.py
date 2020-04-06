import os
from itertools import chain
from setuptools import find_packages, setup

version = {}
dir_path = os.path.dirname(os.path.realpath(__file__))
with open(os.path.join(dir_path, "seldon_core/version.py")) as fp:
    exec(fp.read(), version)

# Extra dependencies, with special 'all' key
extras = {"tensorflow": ["tensorflow"], "gcs": ["google-cloud-storage >= 1.16.0"]}
all_extra_deps = chain.from_iterable(extras.values())
extras["all"] = list(set(all_extra_deps))

setup(
    name="seldon-core",
    author="Seldon Technologies Ltd.",
    author_email="hello@seldon.io",
    version=version["__version__"],
    description="Seldon Core client and microservice wrapper",
    url="https://github.com/SeldonIO/seldon-core",
    license="Apache 2.0",
    packages=find_packages(),
    include_package_data=True,
    setup_requires=["pytest-runner"],
    python_requires=">=3.6",
    install_requires=[
        "Flask<2.0.0",
        "Flask-cors<4.0.0",
        "redis<4.0.0",
        "requests<3.0.0",
        "numpy<2.0.0",
        "flatbuffers<2.0.0",
        "protobuf<4.0.0",
        "grpcio<2.0.0",
        "Flask-OpenTracing >= 1.1.0, < 1.2.0",
        "opentracing >= 2.2.0, < 2.3.0",
        "jaeger-client >= 4.1.0, < 4.2.0",
        "grpcio-opentracing >= 1.1.4, < 1.2.0",
        "pyaml<20.0.0",
        "gunicorn >= 19.9.0, < 20.1.0",
        "minio >= 4.0.9, < 6.0.0",
        "azure-storage-blob >= 2.0.1, < 3.0.0",
        "setuptools >= 41.0.0",
        "prometheus_client >= 0.7.1, < 0.8.0",
    ],
    tests_require=["pytest<6.0.0", "pytest-cov<3.0.0", "Pillow==7.1.1"],
    extras_require=extras,
    test_suite="tests",
    entry_points={
        "console_scripts": [
            "seldon-core-microservice = seldon_core.microservice:main",
            "seldon-core-tester = seldon_core.microservice_tester:main",
            "seldon-core-microservice-tester = seldon_core.microservice_tester:main",
            "seldon-core-api-tester = seldon_core.api_tester:main",
        ]
    },
    zip_safe=False,
)
