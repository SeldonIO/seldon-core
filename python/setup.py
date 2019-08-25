from setuptools import find_packages, setup
import os

version = {}
dir_path = os.path.dirname(os.path.realpath(__file__))
with open(os.path.join(dir_path, "seldon_core/version.py")) as fp:
    exec(fp.read(), version)

setup(name='seldon-core',
      author='Seldon Technologies Ltd.',
      author_email='hello@seldon.io',
      version=version['__version__'],
      description='Seldon Core client and microservice wrapper',
      url='https://github.com/SeldonIO/seldon-core',
      license='Apache 2.0',
      packages=find_packages(),
      include_package_data=True,
      setup_requires=[
          'pytest-runner'
      ],
      python_requires='>=3.6',
      install_requires=[
          'Flask==1.1.1',
          'Flask-Cors==3.0.8',
          'redis==3.3.7',
          'requests==2.22.0',
          'numpy==1.17.0',
          'flatbuffers==1.11',
          'protobuf==3.9.1',
          'grpcio==1.23.0',
          'tensorflow==1.13.1',
          'Flask-OpenTracing==0.2.0',
          'jaeger-client==3.13.0',
          'opentracing>=1.2.2,<2',
          'jaeger-client==3.13.0',
          'grpcio-opentracing==1.1.4',
          'pyaml==19.4.1',
          'gunicorn>=19.9.0',
          'minio>=4.0.9',
          "google-cloud-storage>=1.16.0",
          "azure-storage-blob>=2.0.1"
      ],
      tests_require=[
          'pytest',
          'pytest-cov'
      ],
      test_suite='tests',
      entry_points={
          'console_scripts': [
              'seldon-core-microservice = seldon_core.microservice:main',
              'seldon-core-tester = seldon_core.microservice_tester:main',
              'seldon-core-microservice-tester = seldon_core.microservice_tester:main',
              'seldon-core-api-tester = seldon_core.api_tester:main',
          ],
      },
      zip_safe=False)
