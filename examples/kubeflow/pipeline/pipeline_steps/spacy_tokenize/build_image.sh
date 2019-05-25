#!/bin/bash

docker build . -t seldon-core-s2i-python3-spacy:0.6 
s2i build . seldon-core-s2i-python3-spacy:0.6 spacy_tokenizer:0.1

