#
# release.py
#

import yaml
from io import StringIO
import pprint
from subprocess import Popen, PIPE
import os
import sys
import argparse
import re
import shutil


def pp(o):
    pprinter = pprint.PrettyPrinter(indent=4)
    pprinter.pprint(o)


def getOpts(cmd_line_args):
    parser = argparse.ArgumentParser(description="Set seldon-core version")
    parser.add_argument("-d", "--debug", action="store_true", help="turn on debugging")
    parser.add_argument("seldon_core_version", help="the version to set")
    opts = parser.parse_args(cmd_line_args)
    return opts


def dict_to_yaml(d):
    return yaml.dump(d, default_flow_style=False)


def yaml_to_dict(yaml_data):
    return yaml.load(StringIO(yaml_data))


def run_command(args, debug=False):
    err, out = None, None
    if debug:
        print("cwd[{}]".format(os.getcwd()))
        print("Executing: " + repr(args))
    p = Popen(args, stdout=PIPE, stderr=PIPE)
    if p.wait() == 0:
        out = p.stdout.read()
        out = out.strip()
    else:
        err = {}
        if p.stderr != None:
            err["stderr"] = p.stderr.read()
            err["stderr"] = err["stderr"].strip()
        if p.stdout != None:
            err["stdout"] = p.stdout.read()
            err["stdout"] = err["stdout"].strip()
    return err, out


def update_pom_file(fpath, seldon_core_version, debug=False):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    comp_dir_path = os.path.dirname(fpath)
    os.chdir(comp_dir_path)
    args = [
        "mvn",
        "versions:set",
        "-DnewVersion={seldon_core_version}".format(**locals()),
    ]
    err, out = run_command(args, debug)
    ##pp(out)
    ##pp(err)
    if err == None:
        print("updated {fpath}".format(**locals()))
    else:
        print("error {fpath}".format(**locals()))
        print(err)


def update_chart_yaml_file(fpath, seldon_core_version, debug=False):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    f = open(fpath)
    yaml_data = f.read()
    f.close()

    d = yaml_to_dict(yaml_data)
    d["version"] = seldon_core_version

    with open(fpath, "w") as f:
        f.write(dict_to_yaml(d))

    print("updated {fpath}".format(**locals()))


def update_operator_values_yaml_file(fpath, seldon_core_version, debug=False):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    f = open(fpath)
    yaml_data = f.read()
    f.close()

    d = yaml_to_dict(yaml_data)
    d["image"]["tag"] = seldon_core_version
    d["engine"]["image"]["tag"] = seldon_core_version

    with open(fpath, "w") as f:
        f.write(dict_to_yaml(d))

    print("updated {fpath}".format(**locals()))


def update_versions_txt(seldon_core_version, debug=False):
    with open("version.txt", "w") as f:
        f.write("{seldon_core_version}\n".format(**locals()))
    print("Updated version.txt")


def update_kustomize_engine_version(seldon_core_version, debug=False):
    args = [
        "sed",
        "-i",
        "s/docker.io\/seldonio\/engine:\(.*\)/docker.io\/seldonio\/engine:{seldon_core_version}/g".format(
            **locals()
        ),
        "operator/config/manager/manager.yaml",
    ]
    err, out = run_command(args, debug)
    # pp(out)
    # pp(err)
    if err == None:
        print("updated kustomize".format(**locals()))
    else:
        print("error updating kustomize".format(**locals()))
        print(err)

def update_kustomize_executor_version(seldon_core_version, debug=False):
    args = [
        "sed",
        "-i",
        "s/seldonio\/seldon-core-executor:\(.*\)/seldonio\/seldon-core-executor:{seldon_core_version}/g".format(
            **locals()
        ),
        "operator/config/manager/manager.yaml",
    ]
    err, out = run_command(args, debug)
    # pp(out)
    # pp(err)
    if err == None:
        print("updated kustomize".format(**locals()))
    else:
        print("error updating kustomize".format(**locals()))
        print(err)

def update_operator_version(seldon_core_version, debug=False):
    fpath = "operator/config/manager/kustomization.yaml"
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        "s/newTag: .*/newTag: {seldon_core_version}/".format(**locals()),
        fpath,
    ]
    err, out = run_command(args, debug)
    if err == None:
        print("updated {fpath}".format(**locals()))
    else:
        print("error updating {fpath}".format(**locals()))
        print(err)


def set_version(
    seldon_core_version,
    pom_files,
    chart_yaml_files,
    operator_values_yaml_file,
    debug=False,
):
    # Normalize file paths
    pom_files_realpaths = [os.path.realpath(x) for x in pom_files]
    chart_yaml_file_realpaths = [os.path.realpath(x) for x in chart_yaml_files]
    operator_values_yaml_file_realpath = (
        os.path.realpath(operator_values_yaml_file)
        if operator_values_yaml_file != None
        else None
    )

    # Update kustomize
    update_kustomize_engine_version(seldon_core_version, debug)
    update_kustomize_executor_version(seldon_core_version, debug)

    # Update operator version
    update_operator_version(seldon_core_version, debug)

    # Update top level versions.txt
    update_versions_txt(seldon_core_version, debug)

    # update the pom files
    for fpath in pom_files_realpaths:
        update_pom_file(fpath, seldon_core_version, debug)

    # update the helm chart files
    for chart_yaml_file_realpath in chart_yaml_file_realpaths:
        update_chart_yaml_file(chart_yaml_file_realpath, seldon_core_version, debug)

    # update the operator helm values file
    if operator_values_yaml_file != None:
        update_operator_values_yaml_file(
            operator_values_yaml_file_realpath, seldon_core_version, debug
        )


def main(argv):
    POM_FILES = ["engine/pom.xml"]
    CHART_YAML_FILES = [
        "helm-charts/seldon-core-operator/Chart.yaml",
        "helm-charts/seldon-core-analytics/Chart.yaml",
    ]
    OPERATOR_VALUES_YAML_FILE = "helm-charts/seldon-core-operator/values.yaml"

    opts = getOpts(argv[1:])
    if opts.debug:
        pp(opts)
    set_version(
        opts.seldon_core_version,
        POM_FILES,
        CHART_YAML_FILES,
        OPERATOR_VALUES_YAML_FILE,
        opts.debug,
    )
    print("done")


if __name__ == "__main__":
    main(sys.argv)
