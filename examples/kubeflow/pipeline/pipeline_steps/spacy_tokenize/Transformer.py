import spacy
import numpy as np
import logging

# from spacy.cli import download
# import importlib
# download("en_core_web_sm")
# importlib.reload(spacy)

nlp = spacy.load('en_core_web_sm', parser=False, entity=False)
    
class Transformer():
    __symbols = set("!$%^&*()_+|~-=`{}[]:\";'<>?,./-")

    def predict(self, X, feature_names=[]):
        logging.warning(X)
        f = np.vectorize(Transformer.transform_to_token, otypes=[object])
        X_tokenized = f(X)
        logging.warning(X_tokenized)
        return X_tokenized

    def fit(self, X, y=None, **fit_params):
        return self
    
    @staticmethod
    def transform_to_token(text):
        str_text = str(text)
        doc = nlp(str_text, disable=['parser', 'tagger', 'ner'])
        tokens = []
        for token in doc:
            if token.like_url:
                clean_token = "URL"
            else:
                clean_token = token.lemma_.lower().strip()
                if len(clean_token) < 1 or clean_token in \
                        Transformer.__symbols: 
                    continue
            tokens.append(clean_token)
        return tokens

