import os
import re

# Directory containing the files to be modified
directory = 'massmarket_hash_event/'

# Identify all pairs of .py and .pyi files in the directory
files_to_modify = []
for filename in os.listdir(directory):
    if filename.endswith('.py') and filename != "__init__.py":
        base_name = filename[:-3]
        py_file = os.path.join(directory, f'{base_name}.py')
        pyi_file = os.path.join(directory, f'{base_name}.pyi')
        if os.path.exists(py_file) and os.path.exists(pyi_file):
            files_to_modify.append((py_file, pyi_file))

def update_imports(file_path):
    with open(file_path, 'r') as file:
        content = file.read()

    # Pattern to match import statements
    # TODO: might need to make sure we are only patching _our_ proto.py imports
    pattern = re.compile(r'^\s*import\s+([\w_]+)\s+as\s+([\w_]+)', re.MULTILINE)

    def replace_import(match):
        module, alias = match.groups()
        return f'from massmarket_hash_event import {module} as {alias}'

    # Update the content with the new import statements
    updated_content = pattern.sub(replace_import, content)

    with open(file_path, 'w') as file:
        file.write(updated_content)

    print(f'Updated imports in {file_path}')

for py_file, pyi_file in files_to_modify:
    update_imports(py_file)
    update_imports(pyi_file)
