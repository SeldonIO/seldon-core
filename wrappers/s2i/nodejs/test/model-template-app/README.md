## NodeJs tensorflow example

This model example takes an input of 10 different features and predicts a out for the same.
For the training part it uses a random normally distributed input set of 100 rows i.e a data set of [100,10] and trains it for another random normally distributed data set of size [100,1].

For every prediction the model expects a dataset of dimension [r,10] where r is the num of input rows to be predicted.

### Pre-requisites

- node(version>=8.11.0)
- npm or yarn

### Installing Dependencies

```
npm install
```

### Traing the model

```
npm start
```

### Running a prediction

```
npm run predict
```
