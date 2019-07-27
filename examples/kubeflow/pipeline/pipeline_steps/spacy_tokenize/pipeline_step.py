import click
import dill
try:
    # Running for tests
    from .Transformer import Transformer
except:
    # Running from CLI
    from Transformer import Transformer

@click.command()
@click.option('--in-path', default="/mnt/clean_text.data")
@click.option('--out-path', default="/mnt/tokenized_text.data")
def run_pipeline(in_path, out_path):
    spacy_transformer = Transformer()
    with open(in_path, 'rb') as in_f:
        x = dill.load(in_f)
    y = spacy_transformer.predict(x)
    with open(out_path, "wb") as out_f:
        dill.dump(y, out_f)

if __name__ == "__main__":
    run_pipeline()

