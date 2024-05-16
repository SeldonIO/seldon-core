export const seldonObjectType = {
  MODEL: Symbol("Model.mlops.seldon.io"),
  PIPELINE: Symbol("Pipeline.mlops.seldon.io"),
  EXPERIMENT: Symbol("Experiment.mlops.seldon.io")
};

export const seldonOpType = {
  CREATE: Symbol("Create"),
  UPDATE: Symbol("Update"),
  DELETE: Symbol("Delete"),
}
