class StreamingModel:
    def __init__(self):
        print("INITIALIZING STREAMINGMODEL")

    def predict(self, data, names=[], meta=[]):
        print(f"Inside predict: data [{data}] names [{names}] meta [{meta}]")
        return data
