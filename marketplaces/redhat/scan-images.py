from subprocess import Popen, PIPE
import os

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

def scan_images(debug=False):
    paths = [
    "operator",
    "executor",
    "engine",
    "examples/models/mean_classifier",
    "components/alibi-detect-server",
    "components/seldon-request-logger",
    "servers/sklearnserver",
    "servers/mlflowserver",
    "servers/xgboostserver",
    "servers/tfserving_proxy",
    "components/alibi-explain-server",
    "components/storage-initializer",
    "servers/tfserving",
    ]

    for path in paths:
        args = [
            "make",
            "-C",
            "../../"+path,
            "redhat-image-scan"
        ]
        err, out = run_command(args, debug)
        if err == None:
            print("updated {path}".format(**locals()))
        else:
            try:
                errStr = str(err["stderr"])
                if errStr.index("The image tag you are pushing already exists.") > 0:
                    print(f"Warning: Image already exists for {path}.")
            except ValueError:
                print("error updating {path}".format(**locals()))
                print(err)


if __name__ == "__main__":
    with open('../../version.txt', 'r') as file:
        version = file.read()
        print(f"Will scan and upload images for {version}")
    scan_images()
