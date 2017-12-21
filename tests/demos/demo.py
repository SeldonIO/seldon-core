import requests
import json
from requests.auth import HTTPBasicAuth
import numpy as np
import threading
import time

CORE_ROOT = "../.."

TESTS_DIR = CORE_ROOT + "/tests"
CM_DIR = CORE_ROOT + "/cluster-manager"

PRED_UNIT_TMP = './pred_unit_template.json'
DEPL_TMP = './depl_template.json'
DEFAULT_MODEL_IMAGE = "gcr.io/seldon-priv/mean_classifier:0.8"

def get_config(tests_dir=TESTS_DIR,cm_dir=CM_DIR):
    
    with open(TESTS_DIR+"/CLUSTER_MANAGER_ENDPOINT",'r') as f:
        cm_endpoint = f.read()[:-1]


    cm_client_secret = "mysecret"

    with open(TESTS_DIR+"/API_ENDPOINT",'r') as f:
        api_endpoint  = f.read()[:-1]

    config = dict(
        cm_endpoint=cm_endpoint,
        cm_client_secret=cm_client_secret,
        api_endpoint = api_endpoint,
        )
    
    return config

def get_template(path):
    template = json.load(open(path,'r'))
    return template

def _to_dict(list_arg):
    if list_arg is None:
        return {}
    else:
        return {i:v for i,v in enumerate(list_arg)}
    
def build_deployment(n_models,router_image=None,name="Deployment",id="depl_id_0",key="key",secret="secret",project_name="MAB_demos",model_images=None,model_names=None):
    
    model_names = _to_dict(model_names)
    model_images = _to_dict(model_images)
    
    deployment = get_template(DEPL_TMP)
    deployment["id"] = id
    deployment["uniqueName"] = name + '_' + id
    deployment["name"] = "deployment_"+name
    deployment["oauth_key"] = key
    deployment["oauth_secret"] = secret
    deployment["predictor"]["imagePullSecrets"] = ["gcr-seldon-priv"]
    deployment["predictor"]["name"] = "predictor_"+name
    deployment["annotations"]["project_name"] = project_name
    deployment["predictor"]["annotations"]["project_name"] = project_name
    
    
    
    router = get_template(PRED_UNIT_TMP)
    router["children"] = [str(i+1) for i in range(n_models)]
    router["name"] = "router"
    router["type"] = "ROUTER"
    router["id"] = "0"
    if router_image is not None:
        # MAB
        router["cluster_resources"]["image"] = router_image.split(':')[0]
        router["cluster_resources"]["version"] = router_image.split(':')[1]
        router["cluster_resources"]["id"] = "2"
        router["subtype"] = "MICROSERVICE"
        router["parameters"] = [
            {"name":"router_id", "type":"STRING","value":"0"},
            {"name":"deployment_id","type":"STRING","value":id},
            {"name":"n_branches","type":"INT","value":n_models}
        ]
    else:
        # AB test
        router["cluster_resources"] = {}
        router["subtype"] = "RANDOM_ABTEST"
        router["parameters"] = [
            {"name":"ratioA", "type":"FLOAT","value":"0.5"}
        ]
        
    deployment["predictor"]["predictiveUnits"].append(router)
    
    for i in range(n_models):
        model = get_template(PRED_UNIT_TMP)
        model["children"] = []
        model["id"] = str(i+1)
        model["type"] = "MODEL"
        model["endpoint"]["type"] = "REST"
        model["subtype"] = "MICROSERVICE"
        model["parameters"] = []
        model["cluster_resources"]["image"] = model_images.get(i,DEFAULT_MODEL_IMAGE).split(":")[0]
        model["cluster_resources"]["version"] = model_images.get(i,DEFAULT_MODEL_IMAGE).split(":")[1]
        model["cluster_resources"]["id"] = str(i+2)
        model["name"] = model_names.get(i,"model_{}".format(i))
        
        deployment["predictor"]["predictiveUnits"].append(model)
        
    return deployment

def feedback_to_tensor(feedback):
    ndarray = np.array(feedback["response"]["response"]["ndarray"])
    feedback["response"]["response"]["tensor"] = {"shape":ndarray.shape,"values":list(ndarray.ravel())}
    del feedback["response"]["response"]["ndarray"]

class ClientException(Exception):
    pass
    
class OAuthClient(object):
    def __init__(self,endpoint,client_key,client_secret):
        self.endpoint = endpoint
        self.token = requests.post(
            "http://{}:{}@{}/oauth/token".format(client_key,client_secret,endpoint),
            headers={"type":"application/json"},
            data={"grant_type": "client_credentials"}).json()['access_token']
    
    def _request(self,method,url,data=None):
        if data is None:
            data = {}
        request = {
            'GET':requests.get,
            'POST':requests.post,
            'DELETE':requests.delete,
            'PUT':requests.put
        }
        response = request[method](
            "http://"+self.endpoint+url, 
            headers={
                "Content-Type":"application/json",
                "Authorization":"Bearer {}".format(self.token)},
            data=json.dumps(data))
        if response.status_code/100!=2:
            raise ClientException(response.text)
        try:
            return response.json()
        except ValueError:
            return response.text
    
    def _get(self,url,data=None):
        return self._request("GET",url,data)
    
    def _post(self,url,data=None):
        return self._request("POST",url,data)
    
    def _delete(self,url,data=None):
        return self._request("DELETE",url,data)
    
    def _put(self,url,data=None):
        return self._request("PUT",url,data)

class ClusterManagerClient(OAuthClient):
    def ping(self):
        return self._get("/ping")
    
    def authping(self):
        return self._get("/api/v1/authping")
    
    def create_deployment(self,deployment):
        response = self._post("/api/v1/deployments",data=deployment)
        return response
    
    def delete_deployment(self,deployment_id):
        response= self._delete("/api/v1/deployments/{}".format(deployment_id))
        return response

class APIFrontEndClient(OAuthClient):
    def ping(self):
        return self._get("/ping")
    
    def predictions(self,request):
        response = self._post("/api/v0.1/predictions",data=request)
        return response
    
    def feedback(self,feedback):
        response = self._post("/api/v0.1/feedback",data=feedback)
        return response

def track_kwargs(func):
    def inner(self,*args,**kwargs):
        self.kwargs = kwargs.keys()
        return func(self,*args,**kwargs)
    return inner
    
class RewardModel(object):
    def get_reward(self,x,y,prediction,routing):
        pass

class BernouilliRouting(RewardModel):
    @track_kwargs
    def __init__(self,probas):
        self.n_models = len(probas)
        self.params = {'proba_model_{}'.format(i):p for i,p in enumerate(probas)}
        self.probas = probas
        
    def get_reward(self,x,y,prediction,routing):
        # probas = [self.params['proba_model_{}'.format(i)] for i in range(self.n_models)]
        
        return float(np.random.random()<self.probas[routing])
    
class XYGenerator(object):
    def __init__(self,**kwargs):
        self.feature_names = []
        
    def sample(self):
        pass

class Dummy2DXY(XYGenerator):
    """
    Features are normally distributed, target is always 0.
    """
    def __init__(self,n_features=2):
        self.n_features = n_features
        self.feature_names = ["feature"+str(i) for i in range(n_features)]
        
    def sample(self):
        return np.random.normal(size=(1,self.n_features)),0

class Client(threading.Thread):
    def __init__(self, api_client, xy_generator, reward_model):
        self.api_client = api_client
        self.xy_generator = xy_generator
        self.reward_model = reward_model
        self.started = False
        self.killed = False
        super(Client,self).__init__()
        
    def stop(self):
        self.started = False
        
    def restart(self):
        self.started = True
        
    def kill(self):
        self.killed = True
        
    def run(self):
        self.started = True
        while True:
            if self.killed:
                break
            if self.started:
                x,y = self.xy_generator.sample()
                request = {
                    'request':{
                        'features':self.xy_generator.feature_names,
                        'ndarray':x.tolist()
                    },
                    'meta':{
                        'puid':'0',
                        'tags':{}
                    }
                }
                response = self.api_client.predictions(request)
                prediction = response['response']['ndarray']
                routing = response['meta']['routing']['0']
                reward = self.reward_model.get_reward(x,y,prediction,routing)

                feedback = {
                    "request":request,
                    "response":response,
                    "reward":reward
                }
                feedback_to_tensor(feedback)
                self.api_client.feedback(feedback)
            else:
                time.sleep(1)
