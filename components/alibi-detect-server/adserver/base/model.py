from typing import List, Dict, Optional, Union

DEFAULT_EVENT_PREFIX = "seldon.ceserver."

class ModelResponse(object):

    def __init__(self, data: Dict, metrics: Optional[List[Dict]]):
        self.data = data
        self.metrics = metrics

class CEModel(object):
    def __init__(self, name: str):
        """
        A CloudEvents model

        Parameters
        ----------
        name
             The name of the model
        """
        self.name = name
        self.ready = False

    def load(self):
        """
        Load the model

        """
        raise NotImplementedError

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> Optional[ModelResponse]:
        """
        Process the event data and return a response

        Parameters
        ----------
        inputs
             Input data
        headers
             Headers from the request

        Returns
        -------
             A response object

        """
        raise NotImplementedError
