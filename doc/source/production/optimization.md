# Inference Optimization

Inference optimization is a complex subject and will depend on your model and use case. This page provides various pieces of advice.


## Custom Python Code Optimization

Using the Seldon python wrapper there are various optimization areas one needs to look at.

### Seldon Protocol Payload Types with REST and gRPC

Depending on whether you want to use REST or gRPC and want to send tensor data the format of the request will have a deserialization/serialization cost in the python wrapper. This is investigated in a [python serialization notebook](../examples/python_serialization.html).

The conclusions are:

  * gRPC is faster than REST
  * tftensor is best for large batch size
  * ndarray with gRPC is bad for large batch size
  * simpler tensor/ndarray is better for small batch size

### KMP_AFFINITY

If you are running inference on Intel CPUs with compatible libraries then correct usage of environment variables for KMP and OMP can be useful. Most of the advice on these subjects usually discusses a singel inference request and how to optimize for low latency. One must be careful when using KMP_AFFINITY when you expect to handle parallel inference requests as they may block in unexpected ways if CPU Affinity is being used. We provide an [example benchmarking notebook](../examples/python_kmp_affinity.html).

There are many resources to loop deeper for your model case. Some we have found are:

   * [Maximize TensorFlow Performance on CPU: Considerations and Recommendations for Inference Workloads](https://software.intel.com/content/www/us/en/develop/articles/maximize-tensorflow-performance-on-cpu-considerations-and-recommendations-for-inference.html)
   * [Tensorflow Issue on KMP_AFFINITY](https://github.com/tensorflow/tensorflow/issues/29354)
   * [Best  Practicesfor  ScalingDeep  LearningTraining  and Inference with TensorFlow* OnIntel® Xeon® Processor Based HPC Infrastructures](https://indico.cern.ch/event/762142/sessions/290684/attachments/1752969/2841011/TensorFlow_Best_Practices_Intel_Xeon_AI-HPC_v1.0_Q3_2018.pdf)
   * [Optimizing BERT model for Intel CPU Cores using ONNX runtime default execution provider](https://cloudblogs.microsoft.com/opensource/2021/03/01/optimizing-bert-model-for-intel-cpu-cores-using-onnx-runtime-default-execution-provider/)
   * [Using Intel OpenMP Thread Affinity for Pinning](https://www.nas.nasa.gov/hecc/support/kb/using-intel-openmp-thread-affinity-for-pinning_285.html)
   * [Consider adjusting OMP_NUM_THREADS environment variable for containerized deployments](https://thoth-station.ninja/j/mkl_threads.html)
   * [AWS Deep Learning Containers](https://docs.aws.amazon.com/deep-learning-containers/latest/devguide/dlc-guide.pdf.pdf)
   * [General Best Practices for Intel® Optimization for TensorFlow](https://github.com/IntelAI/models/blob/master/docs/general/tensorflow/GeneralBestPractices.md)

### gRPC multi-processing

From 1.10.0 release of Seldon Core the python wrapper gRPC server will also respect GUNICORN_NUM_WORKERS and be able to handle parallel gRPC requests.

## Benchmarks

We provide links to various [benchmarking notebooks](../examples/notebooks.html#benchmarking-and-load-tests).

