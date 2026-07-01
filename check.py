import os, re
import json

pattern = re.compile(r'"(?:[^"\\]|\\.)*[\u4e00-\u9fa5]+(?:[^"\\]|\\.)*"')
results = []

for root, dirs, files in os.walk('.'):
    if 'vendor' in dirs: dirs.remove('vendor')
    if '.git' in dirs: dirs.remove('.git')
    for file in files:
        if file.endswith('.go'):
            path = os.path.join(root, file)
            with open(path, 'r', encoding='utf-8') as f:
                for i, line in enumerate(f):
                    for match in pattern.findall(line):
                        results.append({"file": path, "line": i+1, "string": match})

with open('remaining_strings.json', 'w', encoding='utf-8') as out:
    json.dump(results, out, ensure_ascii=False, indent=2)
print(f"Total remaining strings: {len(results)}")
