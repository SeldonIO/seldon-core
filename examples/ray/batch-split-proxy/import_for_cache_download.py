from simpletransformers.model import TransformerModel


if __name__ == "__main__":
    TransformerModel(
        "roberta", "roberta-base", args=({"fp16": False}), use_cuda=False,
    )
