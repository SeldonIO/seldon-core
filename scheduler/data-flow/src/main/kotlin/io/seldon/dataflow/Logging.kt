/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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