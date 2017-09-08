import sys, json;
data = json.load(sys.stdin)
print data
if data["cmstatus"]["code"] == 200:
    predictor = data["deployment_result"]["deployment"]["predictor"]
    if predictor["replicas"] == predictor["replicasReady"]:
        sys.exit(0)
    else:
        sys.exit(1)

