import graphviz
import json

def _populate_graph(dot, root, suffix=''):
    name = root.get("name")
    id = name+suffix
    if root.get("implementation"):
        dot.node(id, label=name, shape="box", style="filled", color="lightgrey")
    else:
        dot.node(id, label=name, shape="box")
    endpoint_type = root.get("endpoint",{}).get("type")
    if endpoint_type is not None:
        dot.node(id+'endpoint', label=endpoint_type)
        dot.edge(id,id+'endpoint')
    for child in root.get("children",[]):
        child_id = _populate_graph(dot,child)
        dot.edge(id, child_id)
    return id

def get_graph(filename,predictor=0):
    deployment = json.load(open(filename,'r'))
    predictors = deployment.get("spec").get("predictors")
    dot = graphviz.Digraph()

    for idx in range(len(predictors)):
        with dot.subgraph(name="cluster_"+str(idx)) as pdot:
            graph = predictors[idx].get("graph")
            _populate_graph(pdot, graph, suffix=str(idx))
            pdot.attr(label="predictor-"+str(idx))
    return dot

