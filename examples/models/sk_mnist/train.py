from sklearn.ensemble import RandomForestClassifier
from sklearn import datasets, metrics
from sklearn.utils import shuffle
from sklearn.datasets import fetch_mldata
from sklearn.externals import joblib

from tensorflow.examples.tutorials.mnist import input_data
mnist = input_data.read_data_sets("MNIST_data/")

mnist_images = mnist.train.images
mnist_labels = mnist.train.labels

# To apply a classifier on this data, we need to flatten the image, to
# turn the data in a (samples, feature) matrix:
n_samples = len(mnist_images)
data = mnist_images.reshape((n_samples, -1))
targets = mnist_labels

data,targets = shuffle(data,targets)
classifier = RandomForestClassifier(n_estimators=30)

# We learn the digits on the first half of the digits
classifier.fit(data[:n_samples // 2], targets[:n_samples // 2])

# Now predict the value of the digit on the second half:
expected = targets[n_samples // 2:]
test_data = data[n_samples // 2:]

print(classifier.score(test_data, expected))

predicted = classifier.predict(data[n_samples // 2:])

print("Classification report for classifier %s:\n%s\n"
                % (classifier, metrics.classification_report(expected, predicted)))
print("Confusion matrix:\n%s" % metrics.confusion_matrix(expected, predicted))

joblib.dump(classifier, 'sk.pkl')

