import json

_type_dict = {
    "INT":int,
    "FLOAT":float,
    "DOUBLE":float,
    "STRING":str
    }

def parse_parameters_string(parameters_string):
    parameters = json.loads(parameters_string)
    parsed_parameters = {}
    for param in parameters:
        name = param.get("name")
        value = param.get("value")
        type_ = param.get("type")
        parsed_parameters[name] = _type_dict[type_](value)
    return parsed_parameters
