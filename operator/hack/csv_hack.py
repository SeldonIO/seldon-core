import yaml
import argparse
import sys

def getOpts(cmd_line_args):
    parser = argparse.ArgumentParser(description="remove csv CRD versions")
    parser.add_argument("path", help="the output path to save result")
    opts = parser.parse_args(cmd_line_args)
    return opts

def remove_versions(filename):
    with open(filename, "r") as stream:
        y = yaml.safe_load(stream)
        del y["spec"]["customresourcedefinitions"]["owned"][2]
        del y["spec"]["customresourcedefinitions"]["owned"][1]
        return y

def str_presenter(dumper, data):
  if len(data.splitlines()) > 1:  # check for multiline string
    return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
  return dumper.represent_scalar('tag:yaml.org,2002:str', data)


def main(argv):
    opts = getOpts(argv[1:])
    print(opts)
    y = remove_versions(opts.path)
    fdata = yaml.dump(y, width=1000, default_flow_style=False, sort_keys=False)
    with open(opts.path, "w") as outfile:
        outfile.write(fdata)


if __name__ == "__main__":
    yaml.add_representer(str, str_presenter)
    main(sys.argv)