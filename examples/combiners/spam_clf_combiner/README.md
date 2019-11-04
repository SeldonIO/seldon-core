
#### Model Deployment using Seldon Transformer + Combiner Component

Consider, we want to predict whether a text message is spam or not. The text data may contain multiple languages as well as each individual text may also have words from different languages. As per your needs, you are only interested in building a model on English language. So, you want to first translate the text message in English and then pass it to model to classify whether it is a spam or not. Moreover, we want to build two classifiers to solve the problem: one using Neural network architecture using keras library and other that uses traditional machine learning classifier from scikit learn. At runtime, we want to run both classifiers on given input and output with the average of the both predictions:

Example:
If classifier 1 outputs the probability of 0.7 for the message to be spam
Classifier 2 outputs the probability of 0.8 for the message to be spam
We will go with the average of both classifier: 0.75 



![Model Pipeline](https://github.com/SandhyaaGopchandani/seldon-core/blob/seldon_component_example/examples/combiners/spam_clf_combiner/seldon_inference_graph.png)



    kubectl apply -f deploy.yaml
