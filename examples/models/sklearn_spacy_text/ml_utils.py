import spacy
import re
import numpy as np
from sklearn.base import TransformerMixin
from html.parser import HTMLParser
import dill
import sys, os

nlp = spacy.load("en_core_web_sm", parser=False, entity=False)


class SpacyTokenTransformer(TransformerMixin):
    __symbols = set("!$%^&*()_+|~-=`{}[]:\";'<>?,./-")

    def transform(self, X, **kwargs):
        f = np.vectorize(SpacyTokenTransformer.transform_to_token, otypes=[object])
        X_tokenized = f(X)
        return X_tokenized

    def fit(self, X, y=None, **fit_params):
        return self

    @staticmethod
    def transform_to_token(text):
        str_text = str(text)
        doc = nlp(str_text, disable=["parser", "tagger", "ner"])
        tokens = []
        for token in doc:
            if token.like_url:
                clean_token = "URL"
            else:
                clean_token = token.lemma_.lower().strip()
                if (
                    len(clean_token) < 1
                    or clean_token in SpacyTokenTransformer.__symbols
                ):
                    continue
            tokens.append(clean_token)
        return tokens


class CleanTextTransformer(TransformerMixin):
    __html_parser = HTMLParser()
    __uplus_pattern = re.compile("\<[uU]\+(?P<digit>[a-zA-Z0-9]+)\>")
    __markup_link_pattern = re.compile("\[(.*)\]\((.*)\)")

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
        html_unescaped = CleanTextTransformer.__html_parser.unescape(decoded)
        html_unescaped = re.sub(r"\r\n", " ", html_unescaped)
        html_unescaped = re.sub(r"\r\r\n", " ", html_unescaped)
        html_unescaped = re.sub(r"\r", " ", html_unescaped)
        html_unescaped = html_unescaped.replace("&gt;", " > ")
        html_unescaped = html_unescaped.replace("&lt;", " < ")
        html_unescaped = html_unescaped.replace("--", " - ")
        html_unescaped = CleanTextTransformer.__uplus_pattern.sub(
            " U\g<digit> ", html_unescaped
        )
        html_unescaped = CleanTextTransformer.__markup_link_pattern.sub(
            " \1 \2 ", html_unescaped
        )
        html_unescaped = html_unescaped.replace("\\", "")
        return html_unescaped
