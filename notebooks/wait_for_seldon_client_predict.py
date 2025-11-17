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
