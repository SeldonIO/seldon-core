import torch


def test_model(model, data):
    probs = model(data)
    y_pred = probs.argmax(1)
    y_true = torch.tensor([5, 0])

    assert torch.all(torch.eq(y_pred, y_true))
