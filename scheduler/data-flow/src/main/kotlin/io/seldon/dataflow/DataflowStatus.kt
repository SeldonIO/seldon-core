package io.seldon.dataflow

import io.klogging.Klogger
import io.klogging.Level
import io.klogging.NoCoLogger
import kotlinx.coroutines.runBlocking

/**
 * An interface designed for returning error or status information from various components.
 *
 * The idea is to leave exception throwing to program logic errors or checking invariants
 * that can not be checked at compile time. Use implementations of this interface as function
 * return values to indicate errors/status updates that require special handling in the code.
 */
interface DataflowStatus {
    var exception : Exception?
    var message : String?

    fun getDescription() : String? {
        val exceptionMsg = this.exception?.message
        return if (exceptionMsg != null) {
            "${this.message} Exception: $exceptionMsg"
        } else {
            this.message
        }
    }

    // log status when logger is in a coroutine
    fun log(logger: Klogger, levelIfNoException: Level) {
        val exceptionMsg = this.exception?.message
        val exceptionCause = this.exception?.cause ?: Exception("")
        val statusMsg = this.message
        if (exceptionMsg != null) {
            runBlocking {
                logger.error(exceptionCause, "$statusMsg, Exception: {exception}", exceptionMsg)
            }
        } else {
            runBlocking {
                logger.log(levelIfNoException, "$statusMsg")
            }
        }
    }

    // leg status when logger is outside coroutines
    fun log(logger: NoCoLogger, levelIfNoException: Level) {
        val exceptionMsg = this.exception?.message
        val exceptionCause = this.exception?.cause ?: Exception("")
        if (exceptionMsg != null) {
            logger.error(exceptionCause, "${this.message}, Exception: {exception}", exceptionMsg)
        } else {
            logger.log(levelIfNoException, "${this.message}")
        }
    }
}

fun <T: DataflowStatus> T.withException(e: Exception) : T {
    this.exception = e
    return this
}

fun <T: DataflowStatus> T.withMessage(msg: String): T {
    this.message = msg
    return this
}

