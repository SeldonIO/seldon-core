import numpy as np

#roughly the v2 format for batch iris
arr = np.array([[6.8, 6.1], [2.8,3.4], [4.8,4.5], [1.4, 1.6]],dtype=object)

print(arr)
print(arr.shape)

#what code current does is toList
print(arr.tolist())

print(np.transpose(arr))
# this comes out as [[6.8,2.8,4.8,1.4],[6.1,3.4,4.5,1.6]] i.e. the ndarray batch representation
#

# that proves transpose could be useful to us as could allow us to reuse code in req logger

#but what we'll get with v2 protocol is actually more like

elems = [np.array([6.8, 6.1]),np.array([2.8,3.4]),np.array([4.8,4.5]), np.array([1.4, 1.6])]

#so how to convert that to a single numpy array?

elemsnumpy = np.stack( elems, axis=0 )

print(elemsnumpy)
#want to be like first line but concatenate does not do it (just flattens)

# try stacking options?
# https://stackoverflow.com/questions/44517809/concatenate-multiple-numpy-arrays-in-one-array
# otherwise have to specify shape by looking at first element and then adding to first dimension

#stack does seem to do it!