import json
import sys

if __name__ == "__main__":

    with open('base.json') as f:
        base = json.load(f)
    with open('paths-ambassador.json') as f:
        paths_external = json.load(f)
    with open('components.json') as f:
        components = json.load(f)
    with open('paths-internal.json') as f:
        paths_internal = json.load(f)
    with open('security.json') as f:
        security = json.load(f)

    #
    # Create engine.oa3.json
    #        
    base["paths"] = paths_external["paths"]
    base["components"] = components["components"]

    with open('engine.oa3.json', 'w') as outfile:
            json.dump(base, outfile, indent=4)

            
    #
    # Create component.swagger.json
    #
    with open('base.json') as f:
        base = json.load(f)
        
    base["paths"] = paths_internal["paths"]
    base["components"] = components["components"]    

    with open('wrapper.oa3.json', 'w') as outfile:
            json.dump(base, outfile, indent=4)

    #
    # Create apife.swagger.json
    #
    with open('base.json') as f:
        base = json.load(f)
        
    base["paths"] = paths_external["paths"]
    base["components"] = components["components"]    
    base["components"]["securitySchemes"] = security["securitySchemes"]
    base["security"] = security["security"]    

    with open('apife.oa3.json', 'w') as outfile:
            json.dump(base, outfile, indent=4)
