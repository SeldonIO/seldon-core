import goslate
import numpy as np

class Translator():
    def __init__(self):
    	self.gs = goslate.Goslate()



    def transform_input(self, text_msg, feature_names=None):

        """the translator logic will go here. This shows a use of simple library. But translator service can be a Machine Learning model itself"""
        translated = self.gs.translate(text_msg[0], 'en')
        return np.array([translated])



# if __name__== "__main__":
#     t = Translator()
#     example = np.array(['Wie läuft dein Tag'])
#     translated = t.transform_input(example)
#     print(translated)




