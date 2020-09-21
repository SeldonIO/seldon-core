import pytest

from http import HTTPStatus
from seldon_core.flask_utils import jsonify


@pytest.mark.parametrize(
    "skip_encoding, response, expected",
    [
        (True, '{"foo": "bar"}', b'{"foo": "bar"}'),
        (False, {"foo": "bar"}, b'{"foo":"bar"}\n'),
        (None, {"foo": "bar"}, b'{"foo":"bar"}\n'),
    ],
)
def test_jsonify(app, skip_encoding, response, expected):
    with app.app_context():
        json_response = None
        if skip_encoding is None:
            json_response = jsonify(response)
        else:
            json_response = jsonify(response, skip_encoding=skip_encoding)

    assert json_response.status_code == HTTPStatus.OK
    assert json_response.get_data() == expected
