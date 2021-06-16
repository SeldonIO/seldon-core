import numpy as np


proba_output = np.array([
				0.1,
				0.9,
				0.6,
				0.4
			])
proba_output.shape = (2, 2)
print(proba_output)

col1 = np.array([6.8, 6.1])
col1.shape = (2,1)
col2 = np.array([2.8,3.4])
col2.shape = (2,1)
col3 = np.array([4.8,4.5])
col3.shape = (2,1)
col4 = np.array([1.4, 1.6])
col4.shape = (2,1)

#roughly the v2 format for batch iris
arr = np.array([col1, col2, col3, col4],dtype=object)

print(arr)
print(arr.shape)

#what code current does is toList
print(arr.tolist())

result = np.transpose(arr).squeeze()
print(result)

#now the proba output
proba = np.array([0.1, 0.9,0.6, 0.4])
proba.shape = (2,2)

print(proba)
#TODO: this isn't quite right either
#would work for proba kind of because we could group afterwards the way we do for ndarray
#but won't work for image
stacked = np.hstack([result, proba])
print(stacked)