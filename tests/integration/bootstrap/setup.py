from http import HTTPStatus

import requests

from fixtures.logger import setup_logger
from fixtures.fusescope import FuseScope

logger = setup_logger(__name__)


def test_setup(fusescope: FuseScope) -> None:
    """Mint (or reuse) the service-account key and confirm the backend is reachable."""
    response = requests.get(f"{fusescope.endpoint}/api/v1/version", timeout=5)
    logger.info("version response: %s", response.status_code)
    assert response.status_code == HTTPStatus.OK


def test_teardown(fusescope: FuseScope) -> None:  # noqa: ARG001 — requesting the fixture drives teardown
    """Delete the cached service account (use with --teardown)."""
