import matplotlib.pyplot as plt
import numpy as np
import itertools
import graphviz
import json
import requests


def plot_confusion_matrix(cm, classes,
                          normalize=False,
                          title='Confusion matrix',
                          cmap=plt.cm.Blues):
    """
    This function prints and plots the confusion matrix.
    Normalization can be applied by setting `normalize=True`.
    """
    if normalize:
        cm = cm.astype('float') / cm.sum(axis=1)[:, np.newaxis]
        print("Normalized confusion matrix")
    else:
        print('Confusion matrix, without normalization')

    print(cm)

    plt.imshow(cm, interpolation='nearest', cmap=cmap)
    plt.title(title)
    plt.colorbar()
    tick_marks = np.arange(len(classes))
    plt.xticks(tick_marks, classes, rotation=45)
    plt.yticks(tick_marks, classes)

    fmt = '.2f' if normalize else 'd'
    thresh = cm.max() / 2.
    for i, j in itertools.product(range(cm.shape[0]), range(cm.shape[1])):
        plt.text(j, i, format(cm[i, j], fmt),
                 horizontalalignment="center",
                 color="white" if cm[i, j] > thresh else "black")

    plt.ylabel('True label')
    plt.xlabel('Predicted label')
    plt.tight_layout()


def rest_request_ambassador(deploymentName, namespace, request, endpoint="localhost:8003"):
    response = requests.post(
        "http://" + endpoint + "/seldon/" + namespace + "/" + deploymentName + "/api/v0.1/predictions",
        json=request)
    return response.json()


def send_feedback_rest(deploymentName, namespace, request, response, reward, truth, endpoint="localhost:8003"):
    feedback = {
        "request": request,
        "response": response,
        "reward": reward,
        "truth": {"data": {"ndarray": truth}}
    }
    response = requests.post(
        "http://" + endpoint + "/seldon/" + namespace + "/" + deploymentName + "/api/v0.1/feedback",
        json=feedback)
    return response.json()


def _populate_graph(dot, root, suffix=''):
    name = root.get("name")
    id = name + suffix
    if root.get("implementation"):
        dot.node(id, label=name, shape="box",
                 style="filled", color="lightgrey")
    else:
        dot.node(id, label=name, shape="box")
    endpoint_type = root.get("endpoint", {}).get("type")
    if endpoint_type is not None:
        dot.node(id + 'endpoint', label=endpoint_type)
        dot.edge(id, id + 'endpoint')
    for child in root.get("children", []):
        child_id = _populate_graph(dot, child)
        dot.edge(id, child_id)
    return id


def get_graph(filename, predictor=0):
    deployment = json.load(open(filename, 'r'))
    predictors = deployment.get("spec").get("predictors")
    dot = graphviz.Digraph()

    for idx in range(len(predictors)):
        with dot.subgraph(name="cluster_" + str(idx)) as pdot:
            graph = predictors[idx].get("graph")
            _populate_graph(pdot, graph, suffix=str(idx))
            pdot.attr(label="predictor-" + str(idx))
    return dot
