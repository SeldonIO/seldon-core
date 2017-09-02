from flask import Flask

from seldon.microservice.predict import predict_blueprint
from seldon.pipeline import PipelineSaver

class Microservices(object):
    """
    Allow creation of predict microservices
    """

    def create_prediction_microservice(self,pipeline,model_name,parameters=None):
        """
        Creates a microservice from a pipeline

        Args:
        - pipeline: string or scikit learn pipeline
        - model_name: string

        Kwargs:
        - parameters: dictionary
        
        """
        print parameters
        app = Flask(__name__)

        if type(pipeline)==str:
            saver = PipelineSaver()
            pipeline = saver.load(pipeline)

        app.config['seldon_pipeline'] = pipeline
        app.config['seldon_model_name'] = model_name
        app.config['seldon_ready'] = True

        app.register_blueprint(predict_blueprint)

        return app
