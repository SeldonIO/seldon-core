from sklearn.externals import joblib
import os

class PipelineSaver(object):

    def load(self,pipeline_folder):
        return joblib.load(pipeline_folder+"/p")

    def save(self,pipeline,pipeline_folder):
        if not os.path.exists(pipeline_folder):
            # logger.info("creating folder %s",pipeline_folder)
            os.makedirs(pipeline_folder)
        filename = pipeline_folder + "/p"
        joblib.dump(pipeline,filename)
