from abc import ABC, abstractmethod
from typing import Dict


class ExplainerModel(ABC):
    def __init__(self, name: str):
        self.name = name
        self.ready = False

    def load(self):
        self.ready = True

    @abstractmethod
    def explain(self, request: Dict) -> Dict:
        """Explain method"""
