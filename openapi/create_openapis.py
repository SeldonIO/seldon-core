import json

if __name__ == "__main__":

    with open('base.json') as f:
        base = json.load(f)
    with open('paths-ambassador.json') as f:
        paths_ambassador = json.load(f)
    with open('components.json') as f:
        components = json.load(f)
    with open('paths-internal.json') as f:
        paths_internal = json.load(f)
    with open('security.json') as f:
        security = json.load(f)

    #
    # Create engine.oas3.json
    #
    base["paths"] = paths_ambassador["paths"]
    base["components"] = components["components"]

    with open('engine.oas3.json', 'w') as outfile:
            json.dump(base, outfile, indent=4)


    #
    # Create component.swagger.json
    #
    with open('base.json') as f:
        base = json.load(f)

    base["paths"] = paths_internal["paths"]
    base["components"] = components["components"]

    with open('wrapper.oas3.json', 'w') as outfile:
            json.dump(base, outfile, indent=4)
