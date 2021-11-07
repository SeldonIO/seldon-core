from typing import Dict
from abc import ABC, abstractmethod


class ExplainerModel(ABC):

    def __init__(self):
        self.ready = False

    def load(self):
        self.ready = True

    @abstractmethod
    def explain(self, request: Dict) -> Dict:
        """Explain method"""
