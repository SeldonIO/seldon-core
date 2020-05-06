env:
	python3 -m venv .env
	./.env/bin/pip install --upgrade pip setuptools
	./.env/bin/pip install -r requirements.txt

train:
	.env/bin/python train_iris.py

model: env train
