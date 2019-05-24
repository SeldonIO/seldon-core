import click
import numpy as np
import dill
import spacy

nlp = spacy.load('en_core_web_sm', parser=False, entity=False)
    
class SpacyTokenTransformer():
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
        doc = nlp(str_text, disable=['parser', 'tagger', 'ner'])
        tokens = []
        for token in doc:
            if token.like_url:
                clean_token = "URL"
            else:
                clean_token = token.lemma_.lower().strip()
                if len(clean_token) < 1 or clean_token in \
                        SpacyTokenTransformer.__symbols: 
                    continue
            tokens.append(clean_token)
        return tokens

@click.command()
@click.option('--in-path', default="/mnt/clean_text.data")
@click.option('--out-path', default="/mnt/tokenized_text.data")
def run_pipeline(in_path, out_path):
    spacy_transformer = SpacyTokenTransformer()
    with open(in_path, 'rb') as in_f:
        x = dill.load(in_f)
    y = spacy_transformer.transform(x)
    with open(out_path, "wb") as out_f:
        dill.dump(y, out_f)

if __name__ == "__main__":
    run_pipeline()

