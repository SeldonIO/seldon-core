import argparse
import glob
import yaml

parser = argparse.ArgumentParser()
parser.add_argument('--input', required=True, help='Input folder')
parser.add_argument('--output', required=True, help='Output folder')
args, _ = parser.parse_known_args()

ROLE_FILE="namespace_role.yaml"
CLUSTERROLE_FILE="cluster_role.yaml"

COMBINED_ROLES_FILE="role.yaml"
COMBINED_ROLES_BINDING_FILE="role_binding.yaml"

OPERATOR_FILE="operator.yaml"
SERVICEACCOUNT_FILE="service_account.yaml"

if __name__ == "__main__":
    exp = args.input + "/*"
    files = glob.glob(exp)

    roleYaml = None
    clusterRoleYaml = None
    clusterRoleBindingYaml = None
    serviceAccountName = "seldon-manager"

    for file in files:
        with open(file, 'r') as stream:
            res = yaml.safe_load(stream)
            kind = res["kind"].lower()
            name = res["metadata"]["name"].lower()

            print("Processing ",file)
            print(kind,name)

            if kind == "deployment" and name == "seldon-controller-manager":
                serviceAccountName = res["spec"]["template"]["spec"]["serviceAccountName"]
                res["metadata"]["namespace"] = "default"
                print("Writing ",OPERATOR_FILE)
                filename = args.output + "/" + OPERATOR_FILE
                fdata = yaml.dump(res, width=1000)
                with open(filename, 'w') as outfile:
                    outfile.write(fdata)
            elif kind == "serviceaccount" and name == "seldon-manager":
                res["metadata"]["namespace"] = "default"
                print("Writing ", SERVICEACCOUNT_FILE)
                filename = args.output + "/" + SERVICEACCOUNT_FILE
                fdata = yaml.dump(res, width=1000)
                with open(filename, 'w') as outfile:
                    outfile.write(fdata)
            elif kind == "role":
                res["metadata"]["namespace"] = "default"
                if roleYaml is None:
                    roleYaml = res
                else:
                    roleYaml["rules"] = roleYaml["rules"] + res["rules"]
            elif kind == "clusterrole":
                if clusterRoleYaml is None:
                    clusterRoleYaml = res
                else:
                    clusterRoleYaml["rules"] = clusterRoleYaml["rules"] + res["rules"]
            elif kind == "clusterrolebinding" and name == "seldon-manager-rolebinding":
                res["roleRef"]["name"] = "seldon-manager"
                res["subjects"][0]["namespace"] = "default"
                filename = args.output + "/" + COMBINED_ROLES_BINDING_FILE
                fdata = yaml.dump(res, width=1000)
                with open(filename, 'w') as outfile:
                    outfile.write(fdata)

    # Write role yaml
    roleYaml["metadata"]["name"] = serviceAccountName
    fdata = yaml.dump(roleYaml, width=1000)
    filename = args.output + "/" + ROLE_FILE
    with open(filename, 'w') as outfile:
        outfile.write(fdata)

    # Write clusterrole yaml
    clusterRoleYaml["metadata"]["name"] = serviceAccountName
    fdata = yaml.dump(clusterRoleYaml, width=1000)
    filename = args.output + "/" + CLUSTERROLE_FILE
    with open(filename, 'w') as outfile:
        outfile.write(fdata)

    # Write combined role yaml
    for rule in roleYaml["rules"]:
        clusterRoleYaml["rules"].append(rule)
    fdata = yaml.dump(clusterRoleYaml, width=1000)
    filename = args.output + "/" + COMBINED_ROLES_FILE
    with open(filename, 'w') as outfile:
        outfile.write(fdata)
