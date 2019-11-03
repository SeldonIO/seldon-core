
#### Model Deployment using Seldon Transformer Component

Consider, we want to predict whether a text message is spam or not. The text data may contain multiple languages as well as each individual text may also have words from different languages. As per your needs, you are only interested in building a model on English language. So, you want to first translate the text message in English and then pass it to model to classify whether it is a spam or not.


![Model Pipeline](https://github.com/SandhyaaGopchandani/seldon-core/blob/seldon_component_example/examples/input_tranformer/seldon_inference_graph.png)



Each component in Seldon Inference graph is a microservice – specifically a docker container. The Seldon graph links those docker containers (in deploy.yaml file) to result in an end output. 


![Graph Explanation](https://github.com/SandhyaaGopchandani/seldon-core/blob/seldon_component_example/examples/input_tranformer/graph_explained.png)


The repository contains the folders for for each docker image needed for seldon graph to work. Please note that the docker image would not work as is. You will want to add correct logic and model files and recreate the docker images for it to wotk. Update deploy.yaml file to add correct reference to the docker images and deploy it to seldon core using:


    kubectl apply -f deploy.yaml
