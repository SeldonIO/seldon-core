#!/usr/bin/env python
import chainer
import numpy as np
from train_mnist import MLP

class MnistClassifier(object):
    def __init__(self, gpu=-1, model_path='result/snapshot_iter_12000', unit=1000):
        self.gpu = gpu

        # Create a same model object as what you used for training
        model = MLP(unit, 10)
        if gpu >= 0:
            model.to_gpu(gpu)

        # Load saved parameters from a NPZ file of the Trainer object
        try:
            chainer.serializers.load_npz(
                model_path, model, path='updater/model:main/predictor/')
        except Exception:
            chainer.serializers.load_npz(
                model_path, model, path='predictor/')

        self.model = model

    def predict(self, X, features_names, meta = None):
        X = np.float32(X)
        if self.gpu >= 0:
            X = chainer.cuda.cupy.asarray(X)
        with chainer.using_config('train', False):
            return self.model(X[None, ...]).array


def main():
    import argparse

    parser = argparse.ArgumentParser(description='Chainer example: MNIST')
    parser.add_argument('--gpu', '-g', type=int, default=-1,
                        help='GPU ID (negative value indicates CPU)')
    parser.add_argument('--snapshot', '-s',
                        default='result/snapshot_iter_12000',
                        help='The path to a saved snapshot (NPZ)')
    parser.add_argument('--unit', '-u', type=int, default=1000,
                        help='Number of units')
    args = parser.parse_args()

    print('GPU: {}'.format(args.gpu))
    print('# unit: {}'.format(args.unit))
    print('')

    # Prepare data
    train, test = chainer.datasets.get_mnist()
    x, answer = test[0]
    x = x.reshape(1, x.size)

    classifier = MnistClassifier(args.gpu, args.snapshot, args.unit)
    res = classifier.predict(x, [])
    prediction = res.argmax()

    print('Prediction:', prediction)
    print('Answer:', answer)


if __name__ == '__main__':
    main()
