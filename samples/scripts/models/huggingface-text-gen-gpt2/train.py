from transformers import (
    GPT2Tokenizer,
    TFGPT2LMHeadModel,
    pipeline,
)


def main() -> None:
    tokenizer = GPT2Tokenizer.from_pretrained("gpt2")
    model = TFGPT2LMHeadModel.from_pretrained("gpt2")

    p = pipeline(task="text-generation", model=model, tokenizer=tokenizer)

    p.save_pretrained("text-generation-model-artefacts")


if __name__ == "__main__":
    print("Building a custom GPT2 HuggingFace model...")
    main()
