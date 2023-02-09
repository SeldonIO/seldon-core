import argparse
import sys

import yaml


def getOpts(cmd_line_args):
    parser = argparse.ArgumentParser(
        description="remove csv CRD versions and update version"
    )
    parser.add_argument("path", help="the output path to save result")
    parser.add_argument("version", help="the release version")
    opts = parser.parse_args(cmd_line_args)
    return opts


def remove_versions(csv):
    del csv["spec"]["customresourcedefinitions"]["owned"][2]
    del csv["spec"]["customresourcedefinitions"]["owned"][1]
    return csv


def update_container_image(csv, version):
    # Update Operator Image
    csv["metadata"]["annotations"]["containerImage"] = (
        "docker.io/seldonio/seldon-core-operator:" + version
    )

    csv["spec"]["install"]["spec"]["deployments"][0]["spec"]["template"]["spec"][
        "containers"
    ][0]["image"] = ("docker.io/seldonio/seldon-core-operator:" + version)

    # Update Executor image
    for n, env in enumerate(
        csv["spec"]["install"]["spec"]["deployments"][0]["spec"]["template"]["spec"][
            "containers"
        ][0]["env"]
    ):
        if env["name"] == "EXECUTOR_CONTAINER_IMAGE_AND_VERSION":
            csv["spec"]["install"]["spec"]["deployments"][0]["spec"]["template"]["spec"][
                "containers"
            ][0]["env"][n]["value"] = (
                "docker.io/seldonio/seldon-core-executor:" + version
            )
    return csv


def str_presenter(dumper, data):
    if len(data.splitlines()) > 1:  # check for multiline string
        return dumper.represent_scalar("tag:yaml.org,2002:str", data, style="|")
    return dumper.represent_scalar("tag:yaml.org,2002:str", data)


def main(argv):
    opts = getOpts(argv[1:])
    print(opts)
    with open(opts.path, "r") as stream:
        csv = yaml.safe_load(stream)
        csv = update_container_image(csv, opts.version)
        fdata = yaml.dump(csv, width=1000, default_flow_style=False, sort_keys=False)
        with open(opts.path, "w") as outfile:
            outfile.write(fdata)


if __name__ == "__main__":
    yaml.add_representer(str, str_presenter)
    main(sys.argv)
