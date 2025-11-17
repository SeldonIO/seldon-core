import os

def test_mode_retry(timeout=60, interval=2, exception=Exception):
    """
    Decorator: In TEST_MODE, run the function with retry_assertion, else run it once.
    Usage:
        @test_mode_retry(timeout=120)
        def my_test(): ...
    """
    def decorator(func):
        def wrapper(*args, **kwargs):
            TEST_MODE = os.environ.get("TEST_MODE", "0") == "1"
            if TEST_MODE:
                return retry_assertion(func, timeout=timeout, interval=interval, exception=exception, *args, **kwargs)
            else:
                return func(*args, **kwargs)
        return wrapper
    return decorator

def retry_assertion(callback, timeout=300, interval=2, exception=Exception, *args, **kwargs):
    """
    Retry a callback until it does not raise the specified exception (default AssertionError), or timeout is reached.
    Args:
        callback: function to call (should raise exception if assertion fails)
        timeout: seconds to wait before raising TimeoutError
        interval: seconds to wait between retries
        exception: exception type to catch and retry on (default AssertionError)
        *args, **kwargs: passed to callback
    Returns:
        The return value of callback if successful
    Raises:
        TimeoutError: if assertion does not pass within timeout
        Exception: if callback raises an exception other than the specified one
    """
    import time
    start = time.time()
    last_exc = None
    while time.time() - start < timeout:
        try:
            return callback(*args, **kwargs)
        except exception as exc:
            last_exc = exc
        time.sleep(interval)
    raise TimeoutError(f"Assertion did not pass within {timeout} seconds. Last exception: {last_exc}")
import time

def wait_for_seldon_client_predict(sc, timeout=60, transport="rest", assert_success=False, **predict_kwargs):
    """
    Wait until SeldonClient can successfully make a prediction.
    Args:
        sc: SeldonClient instance
        timeout: seconds to wait before raising TimeoutError
        transport: "rest" or "grpc" (passed to sc.predict)
        assert_success: if True, assert the response is successful before returning
        **predict_kwargs: extra kwargs to pass to sc.predict
    Returns:
        The last response object from sc.predict
    Raises:
        TimeoutError: if prediction is not successful within timeout
        AssertionError: if assert_success is True and prediction is not successful
    """
    start = time.time()
    last_response = None
    while time.time() - start < timeout:
        try:
            r = sc.predict(transport=transport, **predict_kwargs)
            last_response = r
            if getattr(r, 'success', False):
                if assert_success:
                    assert r.success, f"Prediction not successful: {getattr(r, 'msg', r)}"
                return r
        except Exception:
            pass
        time.sleep(2)
    raise TimeoutError(f"Seldon deployment not ready within timeout. Last response: {last_response}")
