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

export const seldonOpExecStatus = {
  OK: Symbol("Ok"),
  FAIL: Symbol("Control-plane failure"),
  CONCURRENT_OP_FAIL: Symbol("Failure because of concurrent operation in another VU")
}

export function getSeldonObjectCommonName(objType) {
  let objName = objType.description.split(".")[0]
  return {
    one: objName,
    many: objName + "s"
  }
}
