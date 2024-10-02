import os
import re
import subprocess
from typing import List, Tuple

def get_markdown_files(directory: str) -> List[str]:
    """
    Recursively find all markdown files in the given directory and all sub-directories.
    """
    markdown_files = []
    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith('.md'):
                markdown_files.append(os.path.join(root, file))
    return markdown_files

def get_current_git_tag() -> str:
    return subprocess.check_output(['git', 'describe', '--tags', '--abbrev=0']).decode('utf-8').strip()

def update_urls(content: str, new_version: str) -> Tuple[str, int]:
    """
    Update URLs in the content with the new version.
    Returns the updated content and the number of replacements made.
    """
    pattern = r'(https://github\.com/SeldonIO/seldon-core/blob/v)(\d+(?:\.\d+)*)(/.+)'
    
    def replace_version(match):
        prefix = match.group(1)
        old_version = match.group(2)
        suffix = match.group(3)
        return f"{prefix}{new_version[1:]}{suffix}"
    
    updated_content, count = re.subn(pattern, replace_version, content)
    return updated_content, count

def process_file(file_path: str, new_version: str) -> int:
    """
    Process a single markdown file, updating URLs if necessary.
    Returns the number of replacements made.
    """
    with open(file_path, 'r') as file:
        content = file.read()

    updated_content, count = update_urls(content, new_version)

    if count > 0:
        with open(file_path, 'w') as file:
            file.write(updated_content)

    return count

def main():
    codebase_dir = './docs-gb'  # GitBook directory, change if needed
    new_version = get_current_git_tag()

    markdown_files = get_markdown_files(codebase_dir)
    total_replacements = 0

    for file_path in markdown_files:
        replacements = process_file(file_path, new_version)
        if replacements > 0:
            print(f"Updated {replacements} URL(s) in {file_path}")
            total_replacements += replacements

    print(f"\nTotal replacements made: {total_replacements}")
    print(f"New version: {new_version}")

if __name__ == "__main__":
    main()