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
import json


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
    return yaml.load(StringIO(yaml_data), Loader=yaml.FullLoader)


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
    cwd = os.getcwd()
    os.chdir(comp_dir_path)

    MAVEN_REPOSITORY_LOCATION = os.getenv("MAVEN_REPOSITORY_LOCATION")
    if MAVEN_REPOSITORY_LOCATION == None:
        args = [
            "mvn",
            "versions:set",
            "-DnewVersion={seldon_core_version}".format(**locals()),
        ]
    else:
        args = [
            "mvn",
            "versions:set",
            "-DnewVersion={seldon_core_version}".format(**locals()),
            "-Dmaven.repo.local={MAVEN_REPOSITORY_LOCATION}".format(**locals()),
        ]

    err, out = run_command(args, debug)

    if err == None:
        print("updated {fpath}".format(**locals()))
    else:
        print("error {fpath}".format(**locals()))
        print(err)
    os.chdir(cwd)


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


def update_helm_values_yaml_file_default_images(
    fpath, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        "s/version: \(.*\)/version: {seldon_core_version}/".format(**locals()),
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print("updated helm values yaml for default images".format(**locals()))
    else:
        print("error updating helm values yaml for default images".format(**locals()))
        print(err)


def update_operator_values_yaml_file_core_images(
    fpath, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        "s/tag: \(.*\)/tag: {seldon_core_version}/".format(**locals()),
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print("updated operator values yaml for core images".format(**locals()))
    else:
        print("error updating operator values yaml for core images".format(**locals()))
        print(err)


def update_operator_values_yaml_file_storage_initializer(
    fpath, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        "s|seldonio/rclone-storage-initializer:\(.*\)|seldonio/rclone-storage-initializer:{seldon_core_version}|".format(
            **locals()
        ),
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print("updated operator values yaml for storage initializer".format(**locals()))
    else:
        print(
            "error updating operator values yaml for storage initializer".format(
                **locals()
            )
        )
        print(err)


def update_operator_values_yaml_file_prepackaged_images(
    current_seldon_core_version, fpath, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        's/defaultImageVersion: "{current_seldon_core_version}"/defaultImageVersion: "{seldon_core_version}"/'.format(
            **locals()
        ),
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print(
            "updated operator values yaml for prepackaged server images".format(
                **locals()
            )
        )
    else:
        print(
            "error updating operator values yaml for prepackaged server images".format(
                **locals()
            )
        )
        print(err)


def update_operator_values_yaml_file_explainer_image(
    fpath, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        "s|seldonio/alibiexplainer:\(.*\)|seldonio/alibiexplainer:{seldon_core_version}|".format(
            **locals()
        ),
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print(
            "updated operator values yaml for prepackaged server images".format(
                **locals()
            )
        )
    else:
        print(
            "error updating operator values yaml for prepackaged server images".format(
                **locals()
            )
        )
        print(err)


def update_operator_kustomize_prepackaged_images(
    current_seldon_core_version, fpath, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        's/"defaultImageVersion": "{current_seldon_core_version}"/"defaultImageVersion": "{seldon_core_version}"/'.format(
            **locals()
        ),
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print(
            "updated operator kustomize yaml for prepackaged server images".format(
                **locals()
            )
        )
    else:
        print(
            "error updating operator kustomize yaml for prepackaged server images".format(
                **locals()
            )
        )
        print(err)


def update_sklearn_models_version(
    fpath, model_name, current_seldon_core_version, seldon_core_version, debug=False
):
    fpath = os.path.realpath(fpath)
    if debug:
        print("processing [{}]".format(fpath))
    args = [
        "sed",
        "-i",
        f"s|gs://seldon-models/v{current_seldon_core_version}/{model_name}|gs://seldon-models/v{seldon_core_version}/{model_name}|",
        fpath,
    ]
    err, out = run_command(args, debug)

    if err == None:
        print(
            f"updated model uri gs://seldon-models/v:{seldon_core_version}{model_name} in {fpath}"
        )
    else:
        print(
            f"error updating model uri gs://seldon-models/v:{seldon_core_version}{model_name} in {fpath}"
        )
        print(err)


def update_versions_txt(seldon_core_version, debug=False):
    with open("version.txt", "w") as f:
        f.write("{seldon_core_version}\n".format(**locals()))
    print("Updated version.txt")


def get_current_version():
    with open("version.txt", "r") as f:
        version = f.read()
        print("Current version from version.txt:", version)
        return version.strip()


def update_versions_py(seldon_core_version, debug=False):
    # Updating the version in setup.py
    args = [
        "sed",
        "-i",
        's/version="\(.*\)"/version="{seldon_core_version}"/g'.format(**locals()),
        "python/setup.py",
    ]
    err, out = run_command(args, debug)
    # Updating the version in module __version__.py
    with open("python/seldon_core/version.py", "w") as f:
        f.write('__version__ = "{seldon_core_version}"\n'.format(**locals()))
    print("Updated python/seldon_core/version.py")


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


def update_image_metadata_json(seldon_core_version, debug=False):
    paths = [
        "examples/models/mean_classifier/image_metadata.json",
        "servers/tfserving_proxy/image_metadata.json",
        "servers/sklearnserver/sklearnserver/image_metadata.json",
        "servers/mlflowserver/mlflowserver/image_metadata.json",
        "servers/xgboostserver/xgboostserver/image_metadata.json",
    ]
    for path in paths:
        path = os.path.realpath(path)
        if debug:
            print("processing [{}]".format(path))
        with open(path) as json_file:
            data = json.load(json_file)
            for label in data["labels"]:
                if "version" in label:
                    label["version"] = f"{seldon_core_version}"
            with open(path, "w") as outfile:
                json.dump(data, outfile)


def update_dockerfile_label_version(seldon_core_version, debug=False):
    paths = [
        "operator/Dockerfile.redhat",
        "engine/Dockerfile.redhat",
        "executor/Dockerfile.executor",
        "executor/Dockerfile.executor.redhat",
        "servers/tfserving/Dockerfile.redhat",
        "components/alibi-detect-server/Dockerfile",
        "components/storage-initializer/Dockerfile",
        "components/seldon-request-logger/Dockerfile",
        "components/alibi-explain-server/Dockerfile",
    ]
    replaces = [
        f's/version=".*" \\\\/version="{seldon_core_version}" \\\\/',
        f"s/ubi8:.*/ubi8:{seldon_core_version}/",
    ]
    for path in paths:
        for replace in replaces:
            if debug:
                print("processing [{}]".format(path))
            args = [
                "sed",
                "-i",
                replace,
                path,
            ]
            err, out = run_command(args, debug)
            if err == None:
                print("updated {path}".format(**locals()))
            else:
                print("error updating {path}".format(**locals()))
                print(err)


def update_python_wrapper_fixed_versions(seldon_core_version, debug=False):

    args = [
        "./hack/update_python_version.sh",
        "{seldon_core_version}".format(**locals()),
    ]
    err, out = run_command(args, debug)

    if err == None:
        print("Updated python wrapper in matching files".format(**locals()))
    else:
        print("error updating python wrapper in matching files".format(**locals()))
        print(err)


def set_version(
    current_seldon_core_version,
    seldon_core_version,
    pom_files,
    chart_yaml_files,
    operator_values_yaml_file,
    operator_kustomize_yaml_file,
    abtest_yaml_file,
    mab_yaml_file,
    model_uri_updates,
    debug=False,
):
    update_python_wrapper_fixed_versions(seldon_core_version, debug)

    # Normalize file paths
    pom_files_realpaths = [os.path.realpath(x) for x in pom_files]
    chart_yaml_file_realpaths = [os.path.realpath(x) for x in chart_yaml_files]
    operator_values_yaml_file_realpath = (
        os.path.realpath(operator_values_yaml_file)
        if operator_values_yaml_file != None
        else None
    )
    operator_kustomize_yaml_file_realpath = (
        os.path.realpath(operator_kustomize_yaml_file)
        if operator_kustomize_yaml_file != None
        else None
    )
    abtest_yaml_file_realpath = (
        os.path.realpath(abtest_yaml_file) if abtest_yaml_file != None else None
    )
    mab_values_yaml_file_realpath = (
        os.path.realpath(mab_yaml_file) if mab_yaml_file != None else None
    )

    # Update kustomize
    update_kustomize_engine_version(seldon_core_version, debug)
    update_kustomize_executor_version(seldon_core_version, debug)
    #
    # Update operator version
    update_operator_version(seldon_core_version, debug)
    #
    # Update top level versions.txt
    update_versions_txt(seldon_core_version, debug)
    #
    # Update version.py in python/seldon_core
    update_versions_py(seldon_core_version, debug)
    #
    # update the pom files
    for fpath in pom_files_realpaths:
        update_pom_file(fpath, seldon_core_version, debug)

    # update the helm chart files
    for chart_yaml_file_realpath in chart_yaml_file_realpaths:
        update_chart_yaml_file(chart_yaml_file_realpath, seldon_core_version, debug)

    # update the operator helm values file
    if operator_values_yaml_file != None:
        update_operator_values_yaml_file_core_images(
            operator_values_yaml_file_realpath, seldon_core_version, debug
        )

    # update the operator helm values files
    if mab_yaml_file != None:
        update_helm_values_yaml_file_default_images(
            mab_values_yaml_file_realpath, seldon_core_version, debug
        )
    if abtest_yaml_file != None:
        update_helm_values_yaml_file_default_images(
            abtest_yaml_file_realpath, seldon_core_version, debug
        )

    if operator_values_yaml_file != None:
        update_operator_values_yaml_file_prepackaged_images(
            current_seldon_core_version,
            operator_values_yaml_file_realpath,
            seldon_core_version,
            debug,
        )
        update_operator_values_yaml_file_explainer_image(
            operator_values_yaml_file_realpath, seldon_core_version, debug
        )
        update_operator_values_yaml_file_storage_initializer(
            operator_values_yaml_file_realpath, seldon_core_version, debug
        )

    if operator_kustomize_yaml_file != None:
        update_operator_kustomize_prepackaged_images(
            current_seldon_core_version,
            operator_kustomize_yaml_file_realpath,
            seldon_core_version,
            debug,
        )

    # update models' uris
    for model_name, paths in model_uri_updates.items():
        for fpath in paths:
            update_sklearn_models_version(
                fpath, model_name, current_seldon_core_version, seldon_core_version
            )

    # Update image version labels
    update_image_metadata_json(seldon_core_version, debug)
    update_dockerfile_label_version(seldon_core_version, debug)


def main(argv):
    POM_FILES = ["engine/pom.xml"]
    CHART_YAML_FILES = [
        "helm-charts/seldon-core-operator/Chart.yaml",
        "helm-charts/seldon-core-analytics/Chart.yaml",
    ]
    OPERATOR_VALUES_YAML_FILE = "helm-charts/seldon-core-operator/values.yaml"
    OPERATOR_KUSTOMIZE_CONFIGMAP = "operator/config/manager/configmap.yaml"
    AB_VALUES_YAML_FILE = "helm-charts/seldon-abtest/values.yaml"
    MAB_VALUES_YAML_FILE = "helm-charts/seldon-mab/values.yaml"

    MODEL_URI_UPDATES = {
        "sklearn/iris": [
            "servers/sklearnserver/samples/iris.yaml",
            "servers/sklearnserver/samples/iris_custom.yaml",
            "servers/sklearnserver/samples/iris_predict.yaml",
            "testing/benchmarking/automated-benchmark/README.ipynb",
            "testing/scripts/test_benchmark.py",
            "notebooks/server_examples.ipynb",
            "notebooks/resources/istio_shadow.yaml",
            "examples/streaming/knative-eventing/README.ipynb",
            "examples/streaming/knative-eventing/README.md",
            "examples/streaming/knative-eventing/assets/simple-iris-deployment.yaml",
            "examples/security/ssl_requests/README.ipynb",
            "examples/security/ssl_requests/README.md",
            "examples/iter8/progressive_rollout/separate_sdeps/abtest.ipynb",
            "examples/iter8/progressive_rollout/separate_sdeps/baseline.yaml",
            "examples/iter8/progressive_rollout/single_sdep/abtest.ipynb",
            "examples/iter8/progressive_rollout/single_sdep/abtest.yaml",
            "examples/iter8/progressive_rollout/single_sdep/promote-v1.yaml",
            "examples/init_containers/custom_init_container.ipynb",
            "examples/feedback/feedback-metrics-server/README.ipynb",
            "examples/feedback/feedback-metrics-server/README.md",
            "examples/feedback/metrics-server/README.ipynb",
            "examples/feedback/metrics-server/README.md",

            "README.md",
        ],
        "sklearn/moviesentiment": [
            "testing/resources/movies-text-explainer.yaml",
            "notebooks/explainer_examples.ipynb",
            "notebooks/resources/moviesentiment_explainer.yaml",
        ],
    }

    opts = getOpts(argv[1:])
    current_version = get_current_version()
    if opts.debug:
        pp(opts)
    set_version(
        current_version,
        opts.seldon_core_version,
        POM_FILES,
        CHART_YAML_FILES,
        OPERATOR_VALUES_YAML_FILE,
        OPERATOR_KUSTOMIZE_CONFIGMAP,
        AB_VALUES_YAML_FILE,
        MAB_VALUES_YAML_FILE,
        MODEL_URI_UPDATES,
        opts.debug,
    )

    print("done")


if __name__ == "__main__":
    main(sys.argv)
