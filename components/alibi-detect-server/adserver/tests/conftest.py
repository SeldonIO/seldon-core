import pytest


@pytest.fixture(autouse=True)
def rclone(monkeypatch):
    monkeypatch.setenv("RCLONE_CONFIG_GS_TYPE", "google cloud storage")
    monkeypatch.setenv("RCLONE_CONFIG_GS_ANONYMOUS", "true")
    yield
