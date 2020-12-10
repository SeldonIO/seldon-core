import argparse
import yaml
import sys
import copy


def getOpts(cmd_line_args):
    parser = argparse.ArgumentParser(
        description="Create graph OpenAPI schema patch. Takes an input template and output Kustomize path for given number of levels"
    )
    parser.add_argument("template", help="the yaml template for 1 level of the graph")
    parser.add_argument("path", help="the output path to save result")
    parser.add_argument(
        "--levels", help="the number of levels to create", type=int, default=10
    )
    opts = parser.parse_args(cmd_line_args)
    return opts


def expand_tmpl(filename, levels):
    with open(filename, "r") as stream:
        y = yaml.safe_load(stream)
        tmpl = y[0]["value"]
        for i in range(levels):
            child = copy.deepcopy(tmpl)
            tmpl["properties"]["children"]["items"] = child
            tmpl = child
            if i == levels - 1:
                del tmpl["properties"]["children"]
        # add replace for second and third schema version
        y.append(copy.deepcopy(y[0]))
        y[1][
            "path"
        ] = "/spec/versions/1/schema/openAPIV3Schema/properties/spec/properties/predictors/items/properties/graph"
        y.append(copy.deepcopy(y[0]))
        y[2][
            "path"
        ] = "/spec/versions/2/schema/openAPIV3Schema/properties/spec/properties/predictors/items/properties/graph"
        return y


def main(argv):
    opts = getOpts(argv[1:])
    print(opts)
    y = expand_tmpl(opts.template, opts.levels)
    fdata = yaml.dump(y, width=1000)
    with open(opts.path, "w") as outfile:
        outfile.write(fdata)


if __name__ == "__main__":
    main(sys.argv)
