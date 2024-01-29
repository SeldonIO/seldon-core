/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import io.klogging.Level
import io.klogging.config.loggingConfiguration
import io.klogging.rendering.RENDER_ANSI
import io.klogging.sending.STDOUT

object Logging {
    private const val stdoutSink = "stdout"

    fun configure(appLevel: Level = Level.INFO, kafkaLevel: Level = Level.WARN) =
        loggingConfiguration {
            kloggingMinLogLevel(appLevel)
            sink(stdoutSink, RENDER_ANSI, STDOUT)
            logging {
                fromLoggerBase("io.seldon")
                toSink(stdoutSink)
            }
            logging {
                fromMinLevel(kafkaLevel) {
                    fromLoggerBase("org.apache")
                    toSink(stdoutSink)
                }
            }
        }
}