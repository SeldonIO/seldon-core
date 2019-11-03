import numpy as np

class Combiner(object):

    def aggregate(self, Xs, features_names=None):

        return max(Xs, key=lambda item: float(item[0]))



if __name__== "__main__":
    clf1_res = np.array(['0.80', 'Spam'])
    clf2_res = np.array(['0.9959868467126312e-04', 'Spam'])

    example = np.array([clf1_res, clf2_res])
    dc = Combiner()
    res = dc.aggregate(example, features_names=None)
    print(res)




