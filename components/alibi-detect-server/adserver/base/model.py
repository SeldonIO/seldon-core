from typing import List, Dict, Optional, Union

DEFAULT_EVENT_PREFIX = "seldon.ceserver."


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

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> Optional[Dict]:
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
