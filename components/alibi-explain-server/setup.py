# Copyright 2019 kubeflow.org.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from setuptools import setup, find_packages

tests_require = [
    'pytest',
    'pytest-tornasync',
    'mypy',
    "tensorflow>=1.10.1,<1.14",
    "keras-applications==1.0.8"
]

setup(
    name='alibiexplainer',
    version='0.3.0',
    author_email='cc@seldon.io',
    license='../../LICENSE.txt',
    url='https://github.com/SeldonIO/seldon-core/tree/master/components/alibi-explain-server',
    description='Model Explaination Server.',
    long_description=open('README.md').read(),
    python_requires='>=3.6',
    packages=find_packages("alibiexplainer"),
    install_requires=[
        "kfserving>=0.3.0",
        "alibi==0.4.0",
        "scikit-learn== 0.20.3",
        "argparse>=1.4.0",
        "requests>=2.22.0",
        "joblib>=0.13.2",
        "pandas>=0.24.2",
        "numpy>=1.16.3",
        "dill>=0.3.0",
        "spacy>=2.1.4",
        "grpcio>=1.22.0",
        "tensorflow>=1.10.1,<1.14",
        "keras-applications==1.0.8"
    ],
    tests_require=tests_require,
    extras_require={'test': tests_require}
)
