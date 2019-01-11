import sys, yaml, json

items = yaml.load_all(sys.stdin)

items = [x for x in items if x is not None]
jlist = {"kind": "List","apiVersion": "v1","metadata": {},"items": items}
json.dump(jlist, sys.stdout, indent=4)

