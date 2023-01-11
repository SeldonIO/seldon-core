```python
from transformers.pipelines import pipeline
from alibi.explainers import AnchorText
import spacy
from alibi.utils import DistilbertBaseUncased
import numpy as np
```

```python
pp = pipeline(
        "text-classification",
        device=-1,
    )
```

```
No model was supplied, defaulted to distilbert-base-uncased-finetuned-sst-2-english and revision af0f99b (https://huggingface.co/distilbert-base-uncased-finetuned-sst-2-english).
Using a pipeline without specifying a model name and revision in production is not recommended.

```

```python
pp(['hello world'])
```

```
[{'label': 'POSITIVE', 'score': 0.9997522234916687}]

```

```python
import spacy
from alibi.utils import spacy_model

model = 'en_core_web_md'
spacy_model(model=model)
nlp = spacy.load(model)
```

```python
import spacy
#loading the english language small model of spacy
en = spacy.load('en_core_web_md')
stopwords = list(en.Defaults.stop_words)
```

```python
def predict_fn(x):
    r = pp(x)
    res = []
    for j in r:
        if j["label"] == "POSITIVE":
            res.append(1)
        else:
            res.append(0)
    return np.array(res)
```

```python
predict_fn(["cambridge is great","Oxford is awlful"])
```

```
array([1, 0])

```

```python
language_model = DistilbertBaseUncased()
explainer = AnchorText(
    predictor=predict_fn,
    sampling_strategy="language_model",   # use language model to predict the masked words
    language_model=language_model,        # language model to be used
    filling="parallel",                   # just one pass through the transformer
    sample_proba=0.5,                     # probability of masking a word
    frac_mask_templates=0.1,              # fraction of masking templates (smaller value -> faster, less diverse)
    use_proba=True,                       # use words distribution when sampling (if False sample uniform)
    top_n=20,                             # consider the fist 20 most likely words
    temperature=1.0,                      # higher temperature implies more randomness when sampling
    stopwords=stopwords,  # those words will not be sampled
    batch_size_lm=32,                     # language model maximum batch size
)
```

```
Some layers from the model checkpoint at distilbert-base-uncased were not used when initializing TFDistilBertForMaskedLM: ['activation_13']
- This IS expected if you are initializing TFDistilBertForMaskedLM from the checkpoint of a model trained on another task or with another architecture (e.g. initializing a BertForSequenceClassification model from a BertForPreTraining model).
- This IS NOT expected if you are initializing TFDistilBertForMaskedLM from the checkpoint of a model that you expect to be exactly identical (initializing a BertForSequenceClassification model from a BertForSequenceClassification model).
All the layers of TFDistilBertForMaskedLM were initialized from the model checkpoint at distilbert-base-uncased.
If your task is similar to the task the model of the checkpoint was trained on, you can already use TFDistilBertForMaskedLM for predictions without further training.

```

```python
explanation = explainer.explain("a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification", threshold=0.95)
```

```python
explanation
```

```
Explanation(meta={
  'name': 'AnchorText',
  'type': ['blackbox'],
  'explanations': ['local'],
  'params': {
              'seed': 0,
              'filling': 'parallel',
              'sample_proba': 0.5,
              'top_n': 20,
              'temperature': 1.0,
              'use_proba': True,
              'frac_mask_templates': 0.1,
              'batch_size_lm': 32,
              'punctuation': '!"#$%&\'()*+,-./:;<=>?@[\\]^_`{|}~',
              'stopwords': ['made', '’s', 'onto', 'him', 'too', 'sixty', 'my', 'herein', 'had', 'thereby', 'your', 'front', 'to', 'and', 'somewhere', 'thereupon', 'regarding', 'latter', 'along', 'what', 'those', 'between', 'somehow', 'for', 'first', 'mostly', 'various', 'indeed', 'do', 'enough', "n't", 'up', "'m", 'back', 'others', 'whom', 'almost', 'further', '’d', 'it', 'of', 'whither', 'thereafter', 'make', '‘d', 'did', 'twenty', "'ve", 'cannot', 'go', 'her', 'same', 'neither', 'when', 'doing', 'noone', 'well', 'while', 'off', 'everyone', 'perhaps', 'themselves', 'moreover', 'become', 'n’t', 'nowhere', 'something', 'however', 'since', 'namely', '‘re', 'only', 'nothing', 'side', 'either', 'even', 'whereby', 'last', 'thus', 'how', 'amongst', '‘s', 'hereafter', 'afterwards', 'rather', 'whose', 'also', 'else', 'besides', 'next', 'myself', 'because', 'nine', 'about', 'can', 'using', 'within', 'am', 'could', '‘m', 'sometimes', 'are', 'whole', 'toward', 'forty', 'ours', 'no', 'ca', 'meanwhile', 'therein', 'very', 'thence', 'anyway', 'yet', 'does', 'hers', 'least', 'twelve', 'was', 'this', 'our', 'say', 'yourselves', 'much', 'elsewhere', 'should', 'their', '’m', 'anyhow', 'show', 'each', 'get', 'beside', 'whether', 'me', 'via', 'therefore', 'fifteen', 'nor', 'the', 'towards', 'once', 'from', 'wherein', 'nevertheless', 'been', 'be', 'if', 'few', 'amount', 'none', 'alone', 'latterly', 'bottom', 'otherwise', 'he', 'several', 'both', 'used', 'see', 'everything', 'fifty', 'someone', 'at', 'thru', "'s", "'re", 'beforehand', 'done', 'hundred', 'through', 'move', 'part', 'across', 'must', 'throughout', 'below', 'yours', 'may', 'into', 'every', 'n‘t', 'hence', 'many', 'there', 'where', 'becoming', 'now', 'we', 'five', 'whence', 'whoever', 'empty', 'anyone', 'were', 'behind', 'being', 'than', 'around', 'less', 'itself', 'seem', 'unless', 'whenever', 'anywhere', 'before', 'quite', 'together', 'serious', 'then', '’re', 'a', 'wherever', 'hereupon', 'take', 'would', 'above', 'always', 'formerly', 'over', 'all', 'due', 'just', 'them', 'she', 'already', 'until', 'two', 'whereas', 'who', 'whereupon', 'so', 'you', 'one', 'seemed', 'without', 'after', 'as', 'eleven', 'such', '‘ve', 'his', 'please', 'anything', 'its', 'again', 'mine', 'during', 'is', 'hereby', 'with', 'ten', 'not', 'here', 'six', "'ll", 'seeming', 'some', 'still', 'often', 'most', 'everywhere', 'upon', 'they', 'which', '’ve', 'under', 'us', 'any', 'ever', 'why', 'beyond', 'give', 'that', 'an', 'himself', 'became', 'except', 'in', 'other', 'against', 'on', 'but', 'has', 'top', 'though', 'i', 'out', 're', 'becomes', '’ll', 'keep', 'by', 'three', 'or', 'might', 'ourselves', 'herself', 'these', 'four', 'whatever', 'seems', 'per', 'full', 'former', 'yourself', "'d", 'another', '‘ll', 'nobody', 'among', 'never', 'put', 'name', 'down', 'eight', 'sometime', 'have', 'will', 'whereafter', 'third', 'although', 'own', 'call', 'really', 'more'],
              'sample_punctuation': False}
            ,
  'version': '0.8.0'}
, data={
  'anchor': ['vapid', 'opaque'],
  'precision': 1.0,
  'coverage': 0.229,
  'raw': {
           'feature': [5, 3],
           'mean': [0.94, 1.0],
           'precision': [0.94, 1.0],
           'coverage': [0.474, 0.229],
           'examples': [{'covered_true': array(['a visually exquisite but sometimes opaque and emotionally vapid experience of style and imagery',
       'a visually exquisite but ultimately opaque and emotionally vapid experience of style and personality',
       'a visually exquisite but visually opaque and emotionally vapid experience of style and personality',
       'a visually exquisite but somewhat opaque and emotionally vapid experience of style and colour',
       'a visually exquisite but deeply opaque and emotionally vapid experience of style and detail',
       'a visually exquisite but emotionally opaque and emotionally vapid experience of style and imagery',
       'a visually exquisite but sometimes opaque and emotionally vapid experience of style and emotion',
       'a visually exquisite but emotionally opaque and emotionally vapid experience of style and emotion',
       'a visually exquisite but emotionally opaque and emotionally vapid experience of style and texture',
       'a visually exquisite but ultimately opaque and emotionally vapid experience of style and emotion'],
      dtype='<U109'), 'covered_false': array(['a rather exquisite but narratively pleasing and deeply vapid tale of joy and mystification',
       'a rather exquisite but narratively touching and darkly vapid portrait of pain and mystification',
       'a visually exquisite but narratively detailed and delicately vapid story of mystery and mystification'],
      dtype='<U109'), 'uncovered_true': array([], dtype=float64), 'uncovered_false': array([], dtype=float64)}, {'covered_true': array(['a visually exquisite but strangely opaque and ultimately vapid experience of beauty and imagination',
       'a visually exquisite but visually opaque and distinctly vapid experience of loneliness and beauty',
       'a visually exquisite but visually opaque and somewhat vapid experience of mystery and deprivation',
       'a visually exquisite but emotionally opaque and occasionally vapid experience of beauty and violence',
       'a visually exquisite but intensely opaque and darkly vapid experience of death and death',
       'a visually exquisite but visually opaque and potentially vapid experience of nature and violence',
       'a visually exquisite but emotionally opaque and often vapid experience of loneliness and violence',
       'a visually exquisite but nonetheless opaque and strangely vapid experience of love and loneliness',
       'a visually exquisite but highly opaque and curiously vapid experience of nature and mystery',
       'a visually exquisite but strangely opaque and visually vapid experience of love and chaos'],
      dtype='<U105'), 'covered_false': array([], dtype='<U105'), 'uncovered_true': array([], dtype=float64), 'uncovered_false': array([], dtype=float64)}],
           'all_precision': 0,
           'num_preds': 1000000,
           'success': True,
           'positions': [9, 6],
           'names': ['vapid', 'opaque'],
           'instance': 'a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification',
           'instances': ['a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification'],
           'prediction': array([0])}
         }
)

```

```python
from alibi.saving import save_explainer
save_explainer(explainer,"./explainer/data")
```

```python
from alibi.saving import load_explainer
load_explainer(path="./explainer/data", predictor=predict_fn)
```

```
2022-10-20 12:55:49.880324: I tensorflow/core/platform/cpu_feature_guard.cc:193] This TensorFlow binary is optimized with oneAPI Deep Neural Network Library (oneDNN) to use the following CPU instructions in performance-critical operations:  AVX2 AVX_VNNI FMA
To enable them in other operations, rebuild TensorFlow with the appropriate compiler flags.
All model checkpoint layers were used when initializing TFDistilBertForMaskedLM.

All the layers of TFDistilBertForMaskedLM were initialized from the model checkpoint at explainer/data/language_model.
If your task is similar to the task the model of the checkpoint was trained on, you can already use TFDistilBertForMaskedLM for predictions without further training.

```

```
AnchorText(meta={
  'name': 'AnchorText',
  'type': ['blackbox'],
  'explanations': ['local'],
  'params': {
              'seed': 0,
              'filling': 'parallel',
              'sample_proba': 0.5,
              'top_n': 20,
              'temperature': 1.0,
              'use_proba': True,
              'frac_mask_templates': 0.1,
              'batch_size_lm': 32,
              'punctuation': '!"#$%&\'()*+,-./:;<=>?@[\\]^_`{|}~',
              'stopwords': ['made', '’s', 'onto', 'him', 'too', 'sixty', 'my', 'herein', 'had', 'thereby', 'your', 'front', 'to', 'and', 'somewhere', 'thereupon', 'regarding', 'latter', 'along', 'what', 'those', 'between', 'somehow', 'for', 'first', 'mostly', 'various', 'indeed', 'do', 'enough', "n't", 'up', "'m", 'back', 'others', 'whom', 'almost', 'further', '’d', 'it', 'of', 'whither', 'thereafter', 'make', '‘d', 'did', 'twenty', "'ve", 'cannot', 'go', 'her', 'same', 'neither', 'when', 'doing', 'noone', 'well', 'while', 'off', 'everyone', 'perhaps', 'themselves', 'moreover', 'become', 'n’t', 'nowhere', 'something', 'however', 'since', 'namely', '‘re', 'only', 'nothing', 'side', 'either', 'even', 'whereby', 'last', 'thus', 'how', 'amongst', '‘s', 'hereafter', 'afterwards', 'rather', 'whose', 'also', 'else', 'besides', 'next', 'myself', 'because', 'nine', 'about', 'can', 'using', 'within', 'am', 'could', '‘m', 'sometimes', 'are', 'whole', 'toward', 'forty', 'ours', 'no', 'ca', 'meanwhile', 'therein', 'very', 'thence', 'anyway', 'yet', 'does', 'hers', 'least', 'twelve', 'was', 'this', 'our', 'say', 'yourselves', 'much', 'elsewhere', 'should', 'their', '’m', 'anyhow', 'show', 'each', 'get', 'beside', 'whether', 'me', 'via', 'therefore', 'fifteen', 'nor', 'the', 'towards', 'once', 'from', 'wherein', 'nevertheless', 'been', 'be', 'if', 'few', 'amount', 'none', 'alone', 'latterly', 'bottom', 'otherwise', 'he', 'several', 'both', 'used', 'see', 'everything', 'fifty', 'someone', 'at', 'thru', "'s", "'re", 'beforehand', 'done', 'hundred', 'through', 'move', 'part', 'across', 'must', 'throughout', 'below', 'yours', 'may', 'into', 'every', 'n‘t', 'hence', 'many', 'there', 'where', 'becoming', 'now', 'we', 'five', 'whence', 'whoever', 'empty', 'anyone', 'were', 'behind', 'being', 'than', 'around', 'less', 'itself', 'seem', 'unless', 'whenever', 'anywhere', 'before', 'quite', 'together', 'serious', 'then', '’re', 'a', 'wherever', 'hereupon', 'take', 'would', 'above', 'always', 'formerly', 'over', 'all', 'due', 'just', 'them', 'she', 'already', 'until', 'two', 'whereas', 'who', 'whereupon', 'so', 'you', 'one', 'seemed', 'without', 'after', 'as', 'eleven', 'such', '‘ve', 'his', 'please', 'anything', 'its', 'again', 'mine', 'during', 'is', 'hereby', 'with', 'ten', 'not', 'here', 'six', "'ll", 'seeming', 'some', 'still', 'often', 'most', 'everywhere', 'upon', 'they', 'which', '’ve', 'under', 'us', 'any', 'ever', 'why', 'beyond', 'give', 'that', 'an', 'himself', 'became', 'except', 'in', 'other', 'against', 'on', 'but', 'has', 'top', 'though', 'i', 'out', 're', 'becomes', '’ll', 'keep', 'by', 'three', 'or', 'might', 'ourselves', 'herself', 'these', 'four', 'whatever', 'seems', 'per', 'full', 'former', 'yourself', "'d", 'another', '‘ll', 'nobody', 'among', 'never', 'put', 'name', 'down', 'eight', 'sometime', 'have', 'will', 'whereafter', 'third', 'although', 'own', 'call', 'really', 'more'],
              'sample_punctuation': False}
            ,
  'version': '0.8.0'}
)

```

```python

```
