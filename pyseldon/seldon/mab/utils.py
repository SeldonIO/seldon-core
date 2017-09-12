def n_success_failures_from_feedback(feedback):
    n_predictions = feedback.response.response.tensor.shape[0]
    n_success = int(feedback.reward*n_predictions)
    n_failures = n_predictions - n_success
    return n_success, n_failures
