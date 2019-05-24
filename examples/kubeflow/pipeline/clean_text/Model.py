import re 
from html.parser import HTMLParser
import numpy as np

class Model():
    __html_parser = HTMLParser()
    __uplus_pattern = \
        re.compile("\<[uU]\+(?P<digit>[a-zA-Z0-9]+)\>")
    __markup_link_pattern = \
        re.compile("\[(.*)\]\((.*)\)")

    def transform(self, X, **kwargs):
        f = np.vectorize(Model.transform_clean_text)
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
        html_unescaped =Model.\
            __html_parser.unescape(decoded) 
        html_unescaped = re.sub(r"\r\n", " ", html_unescaped)
        html_unescaped = re.sub(r"\r\r\n", " ", html_unescaped)
        html_unescaped = re.sub(r"\r", " ", html_unescaped)
        html_unescaped = html_unescaped.replace("&gt;", " > ")
        html_unescaped = html_unescaped.replace("&lt;", " < ")
        html_unescaped = html_unescaped.replace("--", " - ")
        html_unescaped = Model.__uplus_pattern.sub(
            " U\g<digit> ", html_unescaped)
        html_unescaped = Model.__markup_link_pattern.sub(
            " \1 \2 ", html_unescaped)
        html_unescaped = html_unescaped.replace("\\", "")
        return html_unescaped

