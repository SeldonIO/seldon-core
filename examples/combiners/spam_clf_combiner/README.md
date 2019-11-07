
#### Model Deployment using Seldon Transformer + Combiner Component

Consider, we want to predict whether a text message is spam or not. The text data may contain multiple languages as well as each individual text may also have words from different languages. As per your needs, you are only interested in building a model on English language. So, you want to first translate the text message in English and then pass it to model to classify whether it is a spam or not. Moreover, we want to train two models: one that uses Neural network architecture using keras library and other that uses traditional machine learning classifier from scikit-learn. At runtime, we want to run both mdoels on a given input and output the average of the both predictions:

Example: <br/>
If classifier 1 outputs the probability of 0.7 for the message to be spam <br/>
Classifier 2 outputs the probability of 0.8 for the message to be spam <br/>
The final model using seldon-core graph would output the average probability of both classifiers: 0.75 



![Model Pipeline](https://github.com/SandhyaaGopchandani/seldon-core/blob/seldon_component_example/examples/combiners/spam_clf_combiner/seldon_inference_graph.png)


Each component in Seldon Inference graph is a microservice – specifically a docker container. The Seldon graph links those docker containers (in deploy.yaml file) to result in an end output. 


![Graph Explanation](https://github.com/SandhyaaGopchandani/seldon-core/blob/seldon_component_example/examples/combiners/spam_clf_combiner/graph_explained.png)

The repository contains the folders for for each docker image needed for seldon graph to work. The images used for this example are already published in public dockerhub. So, deploy the model on k8s cluster using:

    kubectl apply -f deploy.yaml
