import pytest
import numpy as np
from collections import Counter
from EpsilonGreedy import EpsilonGreedy

def test_no_branches():
    with pytest.raises(TypeError):
        eg = EpsilonGreedy()

def test_negative_branches():
    with pytest.raises(ValueError):
        eg = EpsilonGreedy(n_branches=-1)

def test_init():
    eg = EpsilonGreedy(n_branches=3, epsilon=0.15, seed=1)

    assert eg.n_branches == 3
    assert eg.epsilon == 0.15
    assert np.array_equal(eg.branch_values, np.zeros_like(eg.branch_values))

def test_statistics():
    eg = EpsilonGreedy(n_branches=3, epsilon=0.1, seed=1, best_branch=0)
    routes = [eg.route(features=None,feature_names=None) for _ in range(100000)]
    counter = Counter(routes)
    print(eg.epsilon)
    assert np.isclose(counter[0]/100000, 1-eg.epsilon, atol=0, rtol=0.01)

