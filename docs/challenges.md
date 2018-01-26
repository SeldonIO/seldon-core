# Challenges

Machine Learning deployment has many challenges which Seldon Core's goals are to solve, these include:

 * Allow a wide range of ML modeling tools to be easily deployed, e.g. Python, R, Spark, and propritary models
 * Launch ML runtime graphs, scale up/down, perform rolling updates
 * Run health checks and ensure recovery of failed components
 * Infrastructure optimization for ML
 * Latency optimization
 * Connect to business apps via various APIs, e.g. REST, gRPC
 * Allow construction of Complex runtime microservice graphs
    * Route requests
    * Transform data
    * Ensembles results
 * Allow various deployment modalities
   * Synchronous
   * Asynchronous
   * Batch
 * Allow Auditing and clear versioning
 * Integrate into Continuous Integration (CI) 
 * Allow Continuous Deployment (CD)
 * Provide Monitoring
    * Base metrics: Accuracy, request latency and throughput
 * Complex metrics: 
   * Concept drift
   * Bias detection
   * Outlier detection
 * Allow for Optimization
   * AB Tests
   * Multi-Armed Bandits

If you see further challenges please add an [Issue](https://github.com/SeldonIO/seldon-core/issues).



