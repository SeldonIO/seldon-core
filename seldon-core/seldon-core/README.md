# Seldon Ksonnet

## Prototypes

There are a selection of prototypes you can use to create your initial deployments

 * **A single model to serve**.
   * ```ks generate seldon-serve-simple-<seldonVersion> mymodel --image <image>```
     * Example: ```ks generate seldon-serve-simple-v1alpha2 mymodel --image seldonio/mock_classifier:1.0```
 * **An A-B test between two models**.
   * ```ks generate seldon-abtest-<seldonVersion> myabtest --imageA <imageA> --imageB <imageB>```
     * Example: ```ks generate seldon-abtest-v1alpha2 myabtest --imageA seldonio/mock_classifier:1.0 --imageB seldonio/mock_classifier:1.0```
 * **A multi-armed bandit between two models**. Allowing you to dynamically push traffic to the best model in real time. For more details see an [e-greedy algorithm example](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/epsilon_greedy_gcp.ipynb).
   * ```ks generate seldon-mab-<seldonVersion> mymab --imageA <imageA> --imageB <imageB>```
     * Example: ```ks generate seldon-mab-v1alpha2 mymab --imageA seldonio/mock_classifier:1.0 --imageB seldonio/mock_classifier:1.0```
 * **An outlier detector for a single model**. See more details on the [default Mahalanobis outlier detection algorithm](https://github.com/SeldonIO/seldon-core/blob/master/examples/transformers/outlier_mahalanobis/outlier_documentation.ipynb).
   * ```ks generate seldon-outlier-detector-<seldonVersion> myout --image <image>```
     * Example: ```ks generate seldon-outlier-detector-v1alpha2 myout --image seldonio/mock_classifier:1.0```

