/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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