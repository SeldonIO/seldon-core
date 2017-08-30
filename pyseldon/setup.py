import os
from setuptools import setup, find_packages
from seldon import __version__

setup(name='seldon',
      version=__version__,
      description='Seldon Python Utilities',
      author='Clive Cox',
      author_email='support@seldon.io',
      license='Apache 2 License',
      setup_requires=['numpy'],
      install_requires=['scikit-learn>=0.17','scipy','numpy',"Flask", "pandas>=0.17","grpc","redis"],
      packages=['seldon', 'seldon.pipeline', 'seldon.microservice',"seldon.mab","seldon.proto"],
      package_dir={},
      package_data={},
      scripts=[],
      )

    
