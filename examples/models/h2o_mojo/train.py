import h2o
h2o.init()
from h2o.estimators.glm import H2OGeneralizedLinearEstimator
path = "http://s3.amazonaws.com/h2o-public-test-data/smalldata/prostate/prostate.csv.zip"
h2o_df = h2o.import_file(path)
h2o_df['CAPSULE'] = h2o_df['CAPSULE'].asfactor()
model = H2OGeneralizedLinearEstimator(family = "binomial")
model.train(y = "CAPSULE",
            x = ["AGE", "RACE", "PSA", "GLEASON"],
            training_frame = h2o_df)
modelfile = model.download_mojo(path="./experiment/", get_genmodel_jar=False)
print("Model saved to " + modelfile)

