from sklearn.externals import joblib
import numpy as np
import os
import logging
from sagemaker_containers.beta.framework import files

logging.basicConfig(format='%(asctime)s %(levelname)s - %(name)s - %(message)s', level=logging.INFO)

logging.getLogger('boto3').setLevel(logging.INFO)
logging.getLogger('s3transfer').setLevel(logging.INFO)
logging.getLogger('botocore').setLevel(logging.WARN)

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)

# Example predict function that extracts a model into the provided model_dir
def model_fn(model_dir):
    model_dir_env = os.environ["SAGEMAKER_MODEL_DIRECTORY"]
    logger.info("Using model directory %s",model_dir_env)
    files.download_and_extract(model_dir_env, "", model_dir)
    clf = joblib.load(os.path.join(model_dir, "model.joblib"))
    return clf

def predict_fn(data, model):
    if len(data.shape) == 1:
        data = np.reshape(data,(1,data.shape[0]))
    return model.predict(data)


