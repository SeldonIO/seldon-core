from bokeh.io import push_notebook, show, output_notebook
from bokeh.layouts import row, column, widgetbox, layout
from bokeh.models.widgets import Dropdown, Button, Select, TextInput
from bokeh.plotting import figure, ColumnDataSource
from bokeh import palettes
from bokeh.models import GraphRenderer, Oval, StaticLayoutProvider
from bokeh.application.handlers import FunctionHandler
from bokeh.application import Application
from bokeh.models.widgets import PreText
from bokeh.models.widgets import Slider

output_notebook()

import random

import logging
# logging.basicConfig()
logging.basicConfig(level=logging.ERROR)

import demo

config = demo.get_config()

print "Connecting to Cluster Manager..."

cm_client = demo.ClusterManagerClient(config["cm_endpoint"],"client",config["cm_client_secret"])

if cm_client.authping()=="authpong":
    print "Cluster Manager Ping Successful"

NUMSTR = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

mabs_options = [
    ("-",None),
    ("abtest","AB Test"),
    ("gcr.io/seldon-priv/mab_epsilon_greedy:0.7","Epsilon Greedy"), 
    ("gcr.io/seldon-priv/mab_thomson_sampling:0.8","Thomson Sampling"), 
    ("gcr.io/seldon-priv/mab_contextual_ts:0.7","Contextual Linear"),
    ("gcr.io/seldon-priv/mab_decaying_ts:0.8","Adaptive TS"),
    ("gcr.io/seldon-priv/mab_contextual_gbnb:0.7","Contextual Binary")]

MODEL_USED = "gcr.io/seldon-priv/meanclassifier:0.8"

def gen_id():
    return str(random.randint(0,1e6))



def find_in_list(l,**kwargs):
    matches = [x for x in l if all([x[k]==v for k,v in kwargs.items()])]
    if len(matches)>0:
        return matches[0]
    else:
        return None

def modify_doc(doc):
    
    state = dict(deployments = [], clients=[], deployments_map=[])
    
    def gen_update_accuracy_param(i_model,deployment_id):
        def inner(attr,old,new):
            i_depl = [i for i,depl in enumerate(state["deployments"]) if depl["id"]==deployment_id][0]
            state["deployments"][i_depl]["model_accuracies"][i_model]= new
        return inner
    
    def gen_delete_button_update(deployment_id, button):
        def inner():
            # index = deployments_map.index(deployment_id)
            index = [i for i,depl in enumerate(state["deployments"]) if depl["id"] == deployment_id][0]
            button.disabled = True
            
            cm_client.delete_deployment(deployment_id)
            
            state["deployments"].pop(index)
            dropdown_deployment.options.pop(index)
            deployments.children.pop(index)
        return inner
    
    def gen_start_client_update(client,button_start,button_stop,button_delete):
        def inner():
            if client.isAlive():
                client.restart()
            else:
                client.start()
            button_start.disabled = True
            button_stop.disabled = False
            button_delete.disabled = True
        return inner
            
    def gen_stop_client_update(client,button_start,button_stop,button_delete):
        def inner():
            client.stop()
            button_start.disabled = False
            button_stop.disabled = True
            button_delete.disabled = False
        return inner

    def gen_delete_client_update(client_id, client, button):
        def inner():
            index = [i for i,c in enumerate(state["clients"]) if c["id"] == client_id][0]
            button.disabled = True
            
            client.kill()
            state["clients"].pop(index)
            clients.children.pop(index)
            
        return inner
    
    def click_deploy():
        mab = dropdown_mab.value
        if mab=="abtest":
            mab = None
        n_models = slider_models.value
        models = [MODEL_USED for dm in range(n_models)]
        depl_name = text_name.value
        
        button_deploy.disabled = True
        dropdown_mab.disabled = True
        slider_models.disabled = True
            
        d_id = gen_id()
        
        state["deployments"].append(
            {
                "id":d_id,
                "name":depl_name,
                "model_accuracies":[0.5 for i in range(n_models)]
            })
        
        deployment_string = demo.build_deployment(
            n_models,
            router_image=mab,
            name=depl_name,
            id=d_id,
            key=d_id,
            secret=d_id,
            project_name="MAB_demos",
            model_images=models,
            model_names=None)
        
        cm_client.create_deployment(deployment_string)

        sliders = [Slider(start=0,end=1,value=0.5,step=0.02,title="Model {} Accuracy".format(NUMSTR[i])) for i in range(n_models)]
        for i,slider in enumerate(sliders):
            slider.on_change('value',gen_update_accuracy_param(i,d_id))
            
        delete_button = Button(label="Delete", button_type="danger")
        
        delete_button.on_click(gen_delete_button_update(d_id, delete_button))
        
        # deployments_map.append(d_id)
        deployments.children.append(
            widgetbox([PreText(text="Deployment "+str(depl_name)),
                    delete_button]+sliders))
        
        dropdown_deployment.options.append((d_id,depl_name))
        if dropdown_deployment.value=='' or dropdown_deployment.value is None:
            dropdown_deployment.value = d_id
        
        button_deploy.disabled = False
        dropdown_mab.disabled = False
        slider_models.disabled = False
        
    def click_create_client():
        
        deployment_id = dropdown_deployment.value
        deployment_name = find_in_list(state["deployments"],id=deployment_id)["name"]
        model_accuracies = find_in_list(state["deployments"],id=deployment_id)["model_accuracies"]
        xy_generator = demo.Dummy2DXY()
        reward_model = demo.BernouilliRouting(model_accuracies)
        
        button_create.disabled = True
            
        c_id = gen_id()
        
        api_client = demo.APIFrontEndClient(
            config["api_endpoint"], 
            client_key = deployment_id, 
            client_secret = deployment_id)
        
        client = demo.Client(api_client,xy_generator,reward_model)

    
        state["clients"].append({
                "id":c_id
            })
    
        start_button = Button(label="Start", button_type="success")
        stop_button = Button(label="Stop", button_type="warning", disabled=True)
        delete_button = Button(label="Delete", button_type="danger")
        
        
        start_button.on_click(gen_start_client_update(client, start_button, stop_button, delete_button))
        stop_button.on_click(gen_stop_client_update(client, start_button, stop_button, delete_button))
        delete_button.on_click(gen_delete_client_update(c_id, client, delete_button))
        
        clients.children.append(
            widgetbox([PreText(text="{} Client".format(deployment_name)),
                        start_button,
                        stop_button,
                        delete_button]))
        
        button_create.disabled = False
        
    # DEPLOYMENTS
    # Controls
    text_name = TextInput(title="Deployment Name")
    dropdown_mab = Select(title="Router", options=mabs_options)
    slider_models = Slider(start=2,end=5,value=2,step=1,title="Number of Models")

    button_deploy = Button(label="Deploy", button_type="success")

    button_deploy.on_click(click_deploy)
    
    all_widgets_depl = [text_name, dropdown_mab, slider_models, button_deploy]
    controls_depl = widgetbox(all_widgets_depl)
    
    deployments = column([])
    
    p1 = figure(tools="", toolbar_location=None,plot_width = 11,plot_height=500)
    
    
    # CLIENTS
    # Controls
    dropdown_deployment = Select(title="Deployment", options=[])
    # dropdown_generator = Select(title="Features Generator", options=xy_gen_options, value= xy_gen_options[0])

    button_create = Button(label="Create Client", button_type="success")
    
    button_create.on_click(click_create_client)
    
    all_widgets_client = [dropdown_deployment, button_create]
    controls_client = widgetbox(all_widgets_client)
    
    clients_map = []
    clients = column([],responsive=True)
    
    p2 = figure(tools="", toolbar_location=None,plot_width = 11,plot_height=1000)
    
    layout = column(row(controls_depl, p1, deployments), row(controls_client, p2,clients))
    doc.add_root(layout)

def get_app():
    handler = FunctionHandler(modify_doc)
    app = Application(handler)
    doc = app.create_document()
    
    return app, doc
