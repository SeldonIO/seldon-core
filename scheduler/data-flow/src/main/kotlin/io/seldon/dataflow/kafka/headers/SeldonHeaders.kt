package io.seldon.dataflow.kafka.headers

object SeldonHeaders {
    const val pipelineName = "pipeline"

    // Remove headers we do not want transferred between topics.
    // These headers are set in MLServer Alibi-Detect and Alibi-Explain runtimes: https://github.com/SeldonIO/mlserver.
    val alibiDiscards = arrayOf(
        "x-seldon-alibi-type",
        "x-seldon-alibi-method",
    )
}