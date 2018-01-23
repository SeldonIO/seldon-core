import os
import shutil
import argparse

def populate_template(template,file_out,**kwargs):
    with open(template,'r') as ftmp:
        with open(file_out,'w') as fout:
            fout.write(ftmp.read().format(**kwargs))

def wrap_model(
        repo,
        model_folder,
        base_image,
        model_name,
        service_type,
        version,
        REST=True,
        out_folder=None,
        force_erase=False,
        persistence=False,
        image_name=None):
    if out_folder is None:
        out_folder = model_folder
    if image_name is None:
        image_name = model_name.lower()
    build_folder = out_folder+'/build'
    if os.path.isdir(build_folder):
        if not force_erase:
            print "Build folder already exists. To force erase, use --force argument"
            exit(0)
        else:
            shutil.rmtree(build_folder)
    shutil.copytree(model_folder,build_folder)
    shutil.copy2("./Makefile",build_folder)
    shutil.copy2('./microservice.py',build_folder)
    shutil.copy2("./persistence.py",build_folder)
    shutil.copy2('./{}_microservice.py'.format(service_type.lower()),build_folder)
    shutil.copy2("./seldon_requirements.txt",build_folder)
    shutil.copytree('./proto',build_folder+'/proto')
    populate_template(
        './Dockerfile.tmp',
        build_folder+'/Dockerfile',
        base_image=base_image,
        model_name=model_name,
        api_type="REST" if REST else "GRPC",
        service_type = service_type,
        persistence = int(persistence)
    )
    populate_template(
        "./build_image.sh.tmp",
        build_folder+"/build_image.sh",
        docker_repo=repo,
        docker_image_name=image_name,
        docker_image_version=version)
    populate_template(
        "./push_image.sh.tmp",
        build_folder+"/push_image.sh",
        docker_repo=repo,
        docker_image_name=image_name,
        docker_image_version=version)

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
    parser.add_argument("--out-folder", type=str, default=None, help="Path to the folder where the pre-wrapped model will be created. Defaults to a 'build' folder in the model directory.")
    parser.add_argument("--service-type", type=str, choices=["MODEL","ROUTER","TRANSFORMER","COMBINER","OUTLIER_DETECTOR"], default="MODEL", help="The type of Seldon API the wrapped model will use. Defaults to MODEL.")
    parser.add_argument("--base-image", type=str, default="python:2", help="The base docker image to inherit from. Defaults to python:2. Caution: this must be a debian based image.")
    parser.add_argument("-f", "--force", action="store_true", help="When this flag is present the script will overwrite the contents of the output folder even if it already exists. By default the script would abort.")
    parser.add_argument("-p", "--persistence", action="store_true", help="Use redis to make the model persistent")
    parser.add_argument("--image-name",type=str,default=None,help="Name to give to the model's docker image. Defaults to the model name in lowercase.")

    args = parser.parse_args()

    wrap_model(
        args.repo,
        args.model_folder,
        args.base_image,
        args.model_name,
        args.service_type,
        args.version,
        REST = not args.grpc,
        out_folder = args.out_folder,
        force_erase = args.force,
        persistence = args.persistence,
        image_name = args.image_name)
