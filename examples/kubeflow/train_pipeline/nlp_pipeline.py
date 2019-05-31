
import kfp.dsl as dsl
import yaml
from kubernetes import client as k8s


@dsl.pipeline(
  name='NLP',
  description='A pipeline demonstrating reproducible steps for NLP'
)
def nlp_pipeline(
        csv_url="https://raw.githubusercontent.com/axsauze/reddit-classification-exploration/master/data/reddit_train.csv",
        csv_encoding="ISO-8859-1",
        features_column="BODY",
        labels_column="REMOVED",
        raw_text_path='/mnt/text.data',
        labels_path='/mnt/labels.data',
        clean_text_path='/mnt/clean.data',
        spacy_tokens_path='/mnt/tokens.data',
        tfidf_vectors_path='/mnt/tfidf.data',
        lr_prediction_path='/mnt/prediction.data',
        tfidf_model_path='/mnt/tfidf.model',
        lr_model_path='/mnt/lr.model',
        lr_c_param=0.1,
        tfidf_max_features=10000,
        tfidf_ngram_range=3,
        batch_size='100'):
    """
    Pipeline 
    """
    vop = dsl.VolumeOp(
      name='my-pvc',
      resource_name="my-pvc",
      modes=["ReadWriteMany"],
      size="1Gi"
    )

    download_step = dsl.ContainerOp(
        name='data_downloader',
        image='data_downloader:0.1',
        command="python",
        arguments=[
            "/microservice/pipeline_step.py",
            "--labels-path", labels_path,
            "--features-path", raw_text_path,
            "--csv-url", csv_url,
            "--csv-encoding", csv_encoding,
            "--features-column", features_column,
            "--labels-column", labels_column
        ],
        pvolumes={"/mnt": vop.volume}
    )

    clean_step = dsl.ContainerOp(
        name='clean_text',
        image='clean_text_transformer:0.1',
        command="python",
        arguments=[
            "/microservice/pipeline_step.py",
            "--in-path", raw_text_path,
            "--out-path", clean_text_path,
        ],
        pvolumes={"/mnt": download_step.pvolume}
    )

    tokenize_step = dsl.ContainerOp(
        name='tokenize',
        image='spacy_tokenizer:0.1',
        command="python",
        arguments=[
            "/microservice/pipeline_step.py",
            "--in-path", clean_text_path,
            "--out-path", spacy_tokens_path,
        ],
        pvolumes={"/mnt": clean_step.pvolume}
    )

    vectorize_step = dsl.ContainerOp(
        name='vectorize',
        image='tfidf_vectorizer:0.1',
        command="python",
        arguments=[
            "/microservice/pipeline_step.py",
            "--in-path", spacy_tokens_path,
            "--out-path", tfidf_vectors_path,
            "--max-features", tfidf_max_features,
            "--ngram-range", tfidf_ngram_range,
            "--action", "train",
            "--model-path", tfidf_model_path,
        ],
        pvolumes={"/mnt": tokenize_step.pvolume}
    )

    predict_step = dsl.ContainerOp(
        name='predictor',
        image='lr_text_classifier:0.1',
        command="python",
        arguments=[
            "/microservice/pipeline_step.py",
            "--in-path", tfidf_vectors_path,
            "--labels-path", labels_path,
            "--out-path", lr_prediction_path,
            "--c-param", lr_c_param,
            "--action", "train",
            "--model-path", lr_model_path,
        ],
        pvolumes={"/mnt": vectorize_step.pvolume}
    )

    try:
        seldon_config = yaml.load(open("../deploy_pipeline/seldon_production_pipeline.yaml"))
    except:
        # If this file is run from the project core directory 
        seldon_config = yaml.load(open("deploy_pipeline/seldon_production_pipeline.yaml"))

    deploy_step = dsl.ResourceOp(
        name="seldondeploy",
        k8s_resource=seldon_config,
        attribute_outputs={"name": "{.metadata.name}"})

    deploy_step.after(predict_step)

if __name__ == '__main__':
  import kfp.compiler as compiler
  compiler.Compiler().compile(nlp_pipeline, __file__ + '.tar.gz')
