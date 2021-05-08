from setuptools import setup, find_packages

tests_require = ["pytest", "pytest-tornasync", "mypy", "requests-mock"]
setup(
    name="adserver",
    version="0.1.0",
    author_email="cc@seldon.io",
    license="https://github.com/seldonio/seldon-core/LICENSE",
    url="https://github.com/seldonio/seldon-core/components/alibi-detect-server",
    description="Alibi Detect Server for use with cloud events",
    python_requires=">3.4",
    packages=find_packages(),
    install_requires=[
        "alibi-detect==0.6.1",
        "kfserving>=0.2.0",
        "argparse >= 1.4.0",
        "numpy >= 1.8.2",
        "cloudevents == 1.2.0",
        "elasticsearch==7.9.1",
        # Fixes #2533
        "google-cloud-core==1.4.1",
        "google-cloud-storage==1.31.2",
        "seldon_core",
        "transformers==3.5.1",
    ],
    tests_require=tests_require,
    extras_require={"test": tests_require},
)
