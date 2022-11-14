huggingface.md: huggingface.ipynb
	jupyter nbconvert \
		huggingface.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output huggingface.md

.PHONY: install-requirements
install-requirements:
	pip install -r requirements.txt
	jupyter nbextension enable --py widgetsnbextension

.PHONY: install-ffmpeg-ubuntu
install-ffmpeg-ubuntu:
	sudo apt-get install ffmpeg
