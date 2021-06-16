import numpy as np



dummyimage = np.ones(shape=[2, 32, 32, 3])
dummyfeat = np.array([6.8, 6.1])

#would love to do below
#but get 'could not broadcast input array from shape (2,32,32,3) into shape (2,)'
#arr2 = np.array([dummyimage, dummyfeat], dtype=object)

arr2 = np.array([dummyimage.tolist(), dummyfeat.tolist()])

print(arr2)
print(arr2.shape)

#
# elems2= [dummyimage, dummyfeat]
#
# elemsnumpy = np.stack( elems2, axis=0 )
#
# print(elemsnumpy)

print(arr2.transpose())