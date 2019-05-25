import click
from click.testing import CliRunner
import dill
import numpy as np

from pipeline_steps.data_downloader.pipeline_step import run_pipeline as download_cli
from pipeline_steps.clean_text.pipeline_step import run_pipeline as clean_cli
from pipeline_steps.spacy_tokenize.pipeline_step import run_pipeline as spacy_cli
from pipeline_steps.tfidf_vectorizer.pipeline_step import run_pipeline as tfidf_cli
from pipeline_steps.lr_text_classifier.pipeline_step import run_pipeline as lr_cli

import sys
sys.path.append("..")

def test_pipeline():

    runner = CliRunner()
    
    with runner.isolated_filesystem():

        # # Test Data Downloader
        # result = runner.invoke(download_cli, [
        #     '--labels-path', 'labels.data',
        #     '--features-path', 'raw_text.data'])

        # with open('raw_text.data', "rb") as f:
        #     assert f
        # with open('labels.data', "rb") as f:
        #     assert f

        # Creating test data
        with open('raw_text.data', 'wb') as f:
            raw_arr = np.array([
                "hello this is a test", 
                "another sentence to process"])
            dill.dump(raw_arr, f)

        # Test Clean text transformer
        result = runner.invoke(clean_cli, [
            '--in-path', 'raw_text.data',
            '--out-path', 'clean_text.data'])

        with open('clean_text.data', "rb") as f:
            clean_arr = dill.load(f)
            assert all(raw_arr == clean_arr)

        # Test spacy tokenizer
        result = runner.invoke(spacy_cli, [
            '--in-path', 'clean_text.data',
            '--out-path', 'tokenized_text.data'])

        with open('tokenized_text.data', "rb") as f:
            tokenized_arr = dill.load(f)
            expected_array = np.array([
                ["hello", "this", "be", "a", "test"],
                ["another", "sentence", "to", "process"]
            ])
            assert all(tokenized_arr == expected_array)

        # Test tfidf vectorizer
        result = runner.invoke(tfidf_cli, [
            '--in-path', 'tokenized_text.data',
            '--out-path', 'tfidf_vectors.data',
            '--max-features', "10",
            '--ngram-range', "3",
            '--action', 'train',
            '--model-path', 'tfidf.model'])

        with open("tfidf_vectors.data", "rb") as f:
            tfidf_vectors = dill.load(f)
            assert tfidf_vectors.shape == (2,10)

        with open("tfidf.model", "rb") as f:
            assert f

        # Test lr model

        with open("labels.data", "wb") as f:
            labels = np.array([0,1])
            dill.dump(labels, f)

        result = runner.invoke(lr_cli, [
            '--in-path', 'tfidf_vectors.data',
            '--labels-path', 'labels.data',
            '--out-path', 'prediction.data',
            '--c-param', "0.1",
            '--action', 'train',
            '--model-path', 'lr_text.model'])

        with open("prediction.data", "rb") as f:
            assert f

        with open("lr_text.model", "rb") as f:
            assert f

