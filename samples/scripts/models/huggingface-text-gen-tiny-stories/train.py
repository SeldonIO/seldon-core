from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    pipeline,
)


def main() -> None:
    tokenizer = AutoTokenizer.from_pretrained("EleutherAI/gpt-neo-125M")
    model = AutoModelForCausalLM.from_pretrained('roneneldan/TinyStories-1M')

    p = pipeline(task="text-generation", model=model, tokenizer=tokenizer)

    p.save_pretrained("text-generation-model-artefacts")


if __name__ == "__main__":
    print("Building a custom Tiny Stories HuggingFace model...")
    main()
