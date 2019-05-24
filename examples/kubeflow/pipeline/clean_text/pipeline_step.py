import re 
from html.parser import HTMLParser
import dill
import click
import numpy as np
import dill

class CleanTextTransformer():
    __html_parser = HTMLParser()
    __uplus_pattern = \
        re.compile("\<[uU]\+(?P<digit>[a-zA-Z0-9]+)\>")
    __markup_link_pattern = \
        re.compile("\[(.*)\]\((.*)\)")

    def transform(self, X, **kwargs):
        f = np.vectorize(CleanTextTransformer.transform_clean_text)
        X_clean = f(X)
        return X_clean

    def fit(self, X, y=None, **fit_params):
        return self
    
    @staticmethod
    def transform_clean_text(raw_text):
        try:
            decoded = raw_text.encode("ISO-8859-1").decode("utf-8")
        except:
            decoded = raw_text.encode("ISO-8859-1").decode("cp1252")
        html_unescaped = CleanTextTransformer.\
            __html_parser.unescape(decoded) 
        html_unescaped = re.sub(r"\r\n", " ", html_unescaped)
        html_unescaped = re.sub(r"\r\r\n", " ", html_unescaped)
        html_unescaped = re.sub(r"\r", " ", html_unescaped)
        html_unescaped = html_unescaped.replace("&gt;", " > ")
        html_unescaped = html_unescaped.replace("&lt;", " < ")
        html_unescaped = html_unescaped.replace("--", " - ")
        html_unescaped = CleanTextTransformer.__uplus_pattern.sub(
            " U\g<digit> ", html_unescaped)
        html_unescaped = CleanTextTransformer.__markup_link_pattern.sub(
            " \1 \2 ", html_unescaped)
        html_unescaped = html_unescaped.replace("\\", "")
        return html_unescaped

@click.command()
@click.option('--in-path', default="/mnt/raw_text.data")
@click.option('--out-path', default="/mnt/clean_text.data")
def run_pipeline(in_path, out_path):
    clean_text_transformer = CleanTextTransformer()
    with open(in_path, 'rb') as in_f:
        x = dill.load(in_f)
    y = clean_text_transformer.transform(x)
    with open(out_path, "wb") as out_f:
        dill.dump(y, out_f)

if __name__ == "__main__":
    run_pipeline()

