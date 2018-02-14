import os
import shutil
import argparse
import jinja2

def populate_template(filename,build_folder,**kwargs):
    with open("./{}.tmp".format(filename),'r') as ftmp:
        with open("{}/{}".format(build_folder,filename),'w') as fout:
            fout.write(jinja2.Template(ftmp.read()).render(kwargs=kwargs,**kwargs))

def wrap_model(
        model_folder,
        build_folder,
        force_erase=False,
        **wrapping_arguments):
    if os.path.isdir(build_folder):
        if not force_erase:
            print("Build folder already exists. To force erase, use --force argument")
            exit(0)
        else:
            shutil.rmtree(build_folder)
    service_type = wrapping_arguments.get("service_type")
    
    shutil.copytree(model_folder,build_folder)
    shutil.copy2("./Makefile",build_folder)
    shutil.copy2('./microservice.py',build_folder)
    shutil.copy2("./persistence.py",build_folder)
    shutil.copy2('./{}_microservice.py'.format(service_type.lower()),build_folder)
    shutil.copy2("./seldon_requirements.txt",build_folder)
    shutil.copytree('./proto',build_folder+'/proto')

    populate_template(
        'Dockerfile',
        build_folder,
        **wrapping_arguments)
    populate_template(
        "build_image.sh",
        build_folder,
        **wrapping_arguments)
    populate_template(
        "push_image.sh",
        build_folder,
        **wrapping_arguments)
    populate_template(
        "README.md",
        build_folder,
        **wrapping_arguments)
     
    # Make the files executable
    st = os.stat(build_folder+"/build_image.sh")
    os.chmod(build_folder+"/build_image.sh", st.st_mode | 0111)
    st = os.stat(build_folder+"/push_image.sh")
    os.chmod(build_folder+"/push_image.sh", st.st_mode | 0111)
    
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Utility script to wrap a python model into a docker build. The scipt will generate build folder that contains a Makefile that can be used to build and publish a Docker Image.")
    
    parser.add_argument("model_folder", type=str, help="Path to the folder that contains the model and all the files to copy in the docker image.")
    parser.add_argument("model_name", type=str, help="Name of the model class and model file without the .py extension")
    parser.add_argument("version", type=str, help="Version string that will be given to the model docker image")
    parser.add_argument("repo", type=str, help="Name of the docker repository to publish the image on")
    parser.add_argument("--grpc", action="store_true", help="When this flag is present the model will be wrapped as a GRPC microservice. By default the model is wrapped as a REST microservice.")
    parser.add_argument("--out-folder", type=str, default=None, help="Path to the folder where the build folder containing the pre-wrapped model will be created. Defaults to the model directory.")
    parser.add_argument("--service-type", type=str, choices=["MODEL","ROUTER","TRANSFORMER","COMBINER","OUTLIER_DETECTOR"], default="MODEL", help="The type of Seldon API the wrapped model will use. Defaults to MODEL.")
    parser.add_argument("--base-image", type=str, default="python:2", help="The base docker image to inherit from. Defaults to python:2. Caution: this must be a debian based image.")
    parser.add_argument("-f", "--force", action="store_true", help="When this flag is present the script will overwrite the contents of the output folder even if it already exists. By default the script would abort.")
    parser.add_argument("-p", "--persistence", action="store_true", help="Use redis to make the model persistent")
    parser.add_argument("--image-name",type=str,default=None,help="Name to give to the model's docker image. Defaults to the model name in lowercase.")

    args = parser.parse_args()
    if args.out_folder is None:
        args.out_folder = args.model_folder
    if args.image_name is None:
        args.image_name = args.model_name.lower()

    wrap_model(
        args.model_folder,
        args.out_folder+"/build",
        force_erase = args.force,
        docker_repo = args.repo,
        base_image = args.base_image,
        model_name = args.model_name,
        api_type = "REST" if not args.grpc else "GRPC",
        service_type = args.service_type,
        docker_image_version = args.version,
        docker_image_name = args.image_name,
        persistence = int(args.persistence))
