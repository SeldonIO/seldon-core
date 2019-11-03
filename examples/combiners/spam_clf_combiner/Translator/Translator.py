import goslate

class Translator():
    def __init__(self):
    	self.gs = goslate.Goslate()



    def transform_input(self, text_msg, feature_names=None):

        """the translator logic will go here. This shows a use of simple library. But translator service can be a Machine Learning model itself"""
        return self.gs.translate(text_msg, 'en')



# if __name__== "__main__":
#     t = Translator()
#     translated = t.transform_input('Wie läuft dein Tag')
#     print(translated)




