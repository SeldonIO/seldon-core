import numpy as np
import os
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline
from sklearn.externals import joblib
from sklearn import datasets

def main():
    clf = LogisticRegression()
    p = Pipeline([('clf', clf)])
    print 'Training model...'
    p.fit(X, y)
    print 'Model trained!'

    filename_p = 'IrisClassifier.sav'
    print 'Saving model in %s' % filename_p
    joblib.dump(p, filename_p)
    print 'Model saved!'
    
if __name__ == "__main__":
    print 'Loading iris data set...'
    iris = datasets.load_iris()
    X, y = iris.data, iris.target
    print 'Dataset loaded!'
    main()
