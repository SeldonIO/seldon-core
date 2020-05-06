from typing import Dict, List


class RequestHandler(object):
    def __init__(self, request: Dict):
        self.request = request

    def validate(self):
        """
        Validate the request

        """
        raise NotImplementedError

    def extract_request(self) -> List:
        """
        Extract the request

        Returns
        -------
             A list
        """
        raise NotImplementedError
