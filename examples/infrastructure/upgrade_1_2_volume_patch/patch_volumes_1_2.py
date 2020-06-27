#!/usr/bin/env python3

import yaml
import subprocess
import os
import time


def run(cmd: str):
    cmd_arr = cmd.split()
    output = subprocess.Popen(
        cmd_arr, stdout=subprocess.PIPE, stderr=subprocess.STDOUT
    ).communicate()
    output_str = [out.decode() for out in output if out]
    return "\n".join(output_str)


def patch_volumes_seldon_1_2():

    namespaces = run("kubectl get ns -o=name")

    for namespace in namespaces.split():
        namespace = namespace.replace("namespace/", "")
        sdeps_raw = run(f"kubectl get sdep -o yaml -n {namespace}")
        sdeps_dict = yaml.safe_load(sdeps_raw)
        sdep_list = sdeps_dict.get("items")
        if sdep_list:
            for sdep in sdep_list:
                name = sdep.get("metadata", {}).get("name")
                print(f"Processing {name} in namespace {namespace}")
                predictors = sdep.get("spec", {}).get("predictors", [])
                for predictor in predictors:
                    for component_spec in predictor.get("componentSpecs", []):
                        for container in component_spec.get("spec", {}).get(
                            "containers", []
                        ):
                            for volume_mount in container.get("volumeMounts", []):
                                if volume_mount.get("name") == "podinfo":
                                    print("Patching volume")
                                    volume_mount["name"] = "seldon-podinfo"

                with open("seldon_tmp.yaml", "w") as tmp_file:
                    yaml.dump(sdep, tmp_file)
                    run("kubectl apply -f seldon_tmp.yaml")

                print(yaml.dump(sdep))
                os.remove("seldon_tmp.yaml")


if __name__ == "__main__":
    patch_volumes_seldon_1_2()
