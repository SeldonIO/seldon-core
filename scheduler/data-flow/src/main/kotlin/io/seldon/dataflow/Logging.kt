package io.seldon.dataflow

import io.klogging.Level
import io.klogging.config.loggingConfiguration
import io.klogging.rendering.RENDER_ANSI
import io.klogging.sending.STDOUT

object Logging {
    private const val stdoutSink = "stdout"

    fun configure(appLevel: Level = Level.INFO, kafkaLevel: Level = Level.WARN) =
        loggingConfiguration {
            kloggingMinLevel(appLevel)
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