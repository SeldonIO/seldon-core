from setuptools import find_packages, setup
from seldon_core import __version__

setup(name='seldon-core',
      author='Seldon Technologies Ltd.',
      author_email='hello@seldon.io',
      version=__version__,
      description='Seldon Core client and microservice wrapper',
      url='https://github.com/SeldonIO/seldon-core',
      license='Apache 2.0',
      packages=find_packages(),
      include_package_data=True,
      setup_requires=[
          'pytest-runner'
      ],
      install_requires=[
          'flask',
          'flask-cors',
          'redis',
          'tornado',
          'requests',
          'numpy',
          'flatbuffers',
          'protobuf',
          'grpcio',
          'tensorflow'
      ],
      tests_require=[
          'pytest',
      ],
      test_suite='tests',
      entry_points={
          'console_scripts': [
              'seldon-core-microservice = seldon_core.microservice:main',
              'seldon-core-tester = seldon_core.tester:main',
              'seldon-core-api-tester = seldon_core.api_tester:main',
          ],
      },
      zip_safe=False)
