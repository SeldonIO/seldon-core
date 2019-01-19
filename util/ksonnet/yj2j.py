import sys, yaml, json
import argparse


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("input", type=str,help="Input file which contains list of json and yaml from helm template")

    args = parser.parse_args()

    with open(args.input, 'r') as myfile:
        data=myfile.read()

    parts = data.split("---\n# Source")

    resources = []

    for part in parts[1:]:
        name = part.split("\n")[0]
        lines = part.split("\n")[1:]
        jory = "\n".join(lines)
        if name.endswith(".yaml"):
            items = yaml.load_all(jory)
            for item in items:
                resources.append(item)
        else:
            j = json.loads(jory)
            if "kind" in j  and j["kind"] == "List":
                for item in j["items"]:
                    resources.append(item)
            else:
                resources.append(j)
                
    jlist = {"kind": "List","apiVersion": "v1","metadata": {},"items": resources}
    json.dump(jlist, sys.stdout, indent=4)

if __name__ == "__main__":
    main()
