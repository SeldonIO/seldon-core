import os
import sys
import inspect
import importlib
import pkgutil
from pathlib import Path
from types import ModuleType
import docstring_parser

# Configurable values
BASE_PACKAGE = "seldon_core"
PYTHON_SRC = Path("python")
DOCS_DIR = Path("docs-gb/api")

# Ensure the Python source directory is in sys.path
sys.path.insert(0, str(PYTHON_SRC.resolve()))
DOCS_DIR.mkdir(parents=True, exist_ok=True)


def write_md_header(title, level=1):
    return f"{'#' * level} {title}\n\n"


def format_docstring(obj):
    raw = inspect.getdoc(obj)
    if not raw:
        return "*No docstring available.*"
    try:
        parsed = docstring_parser.parse(raw)
    except Exception:
        return raw  # fallback if badly formatted

    output = []
    if parsed.short_description:
        output.append(parsed.short_description)
    if parsed.long_description:
        output.append(parsed.long_description)
    if parsed.params:
        output.append("\n**Parameters:**")
        for p in parsed.params:
            output.append(f"- `{p.arg_name}` ({p.type_name or 'unknown'}): {p.description}")
    if parsed.returns:
        output.append("\n**Returns:**")
        output.append(f"- ({parsed.returns.type_name or 'unknown'}): {parsed.returns.description}")
    return "\n".join(output)


def document_module(mod: ModuleType, module_name: str) -> str:
    lines = [write_md_header(f"Module `{module_name}`")]

    for name, member in inspect.getmembers(mod):
        if name.startswith("_"):
            continue

        if inspect.isclass(member) and member.__module__ == module_name:
            lines.append(write_md_header(f"Class `{name}`", 2))
            lines.append(f"**Description:**\n{format_docstring(member)}\n")
            for meth_name, meth in inspect.getmembers(member, inspect.isfunction):
                if meth_name.startswith("_"):
                    continue
                lines.append(write_md_header(f"Method `{meth_name}`", 3))
                lines.append(f"**Signature:** `{meth_name}{inspect.signature(meth)}`\n\n")
                lines.append(f"**Description:**\n{format_docstring(meth)}\n")

        elif inspect.isfunction(member) and member.__module__ == module_name:
            lines.append(write_md_header(f"Function `{name}`", 2))
            lines.append(f"**Signature:** `{name}{inspect.signature(member)}`\n\n")
            lines.append(f"**Description:**\n{format_docstring(member)}\n")

    return "\n".join(lines)


def walk_package(package_name):
    package = importlib.import_module(package_name)
    yield package_name, package
    for _, name, _ in pkgutil.walk_packages(package.__path__, prefix=f"{package_name}."):
        try:
            submodule = importlib.import_module(name)
            yield name, submodule
        except Exception as e:
            print(f"⚠️ Skipping {name}: {e}")
            continue


def generate_all_docs():
    (DOCS_DIR / "index.md").write_text(write_md_header("API Reference"))

    for modname, module in walk_package(BASE_PACKAGE):
        rel_path = modname.replace(".", "/") + ".md"
        out_path = DOCS_DIR / rel_path
        out_path.parent.mkdir(parents=True, exist_ok=True)
        print(f"📄 Generating: {rel_path}")
        out_path.write_text(document_module(module, modname))

    print(f"\n✅ Docs written to: {DOCS_DIR}/")


if __name__ == "__main__":
    generate_all_docs()
