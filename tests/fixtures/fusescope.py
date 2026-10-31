import os
import time
from dataclasses import dataclass

import pytest
import requests

from fixtures import reuse
from fixtures.logger import setup_logger

logger = setup_logger(__name__)

ENV_ENDPOINT = "FUSESCOPE_ENDPOINT"
ENV_ACCESS_TOKEN = "FUSESCOPE_ACCESS_TOKEN"
ENV_ROOT_EMAIL = "FUSESCOPE_ROOT_EMAIL"
ENV_ROOT_PASSWORD = "FUSESCOPE_ROOT_PASSWORD"

# Defaults match a backend started with SIGNOZ_USER_ROOT_EMAIL / SIGNOZ_USER_ROOT_PASSWORD
# set to these values (see tests/README.md).
DEFAULT_ROOT_EMAIL = "admin@integration.test"
DEFAULT_ROOT_PASSWORD = "password123Z$"

# Service account minted for Terraform. signoz-admin is the backend's managed
# admin role, so it can manage every resource the provider exercises.
SERVICE_ACCOUNT_NAME = "terraform-integration"
SERVICE_ACCOUNT_ROLE = "signoz-admin"


@dataclass(frozen=True)
class FuseScope:
    endpoint: str
    access_token: str
    root_token: str = ""
    service_account_id: str = ""

    def __cache__(self) -> dict:
        return {"endpoint": self.endpoint, "access_token": self.access_token, "root_token": self.root_token, "service_account_id": self.service_account_id}

    def __log__(self) -> str:
        return f"fusescope(endpoint={self.endpoint})"


def _login_as_root(endpoint: str, email: str, password: str, *, ready_timeout: float = 240.0) -> str:
    """Return a bearer access token for the root user, retrying until the backend is ready."""
    deadline = time.time() + ready_timeout
    last = None

    while time.time() < deadline:
        try:
            ctx = requests.get(
                f"{endpoint}/api/v2/sessions/context",
                params={"email": email, "ref": endpoint},
                timeout=10,
            )
            if ctx.status_code == 200 and ctx.json().get("data", {}).get("orgs"):
                org_id = ctx.json()["data"]["orgs"][0]["id"]

                login = requests.post(
                    f"{endpoint}/api/v2/sessions/email_password",
                    json={"email": email, "password": password, "orgId": org_id},
                    timeout=10,
                )
                if login.status_code == 200:
                    logger.info("logged in as root user %s", email)
                    return login.json()["data"]["accessToken"]

                last = (login.status_code, login.text[:200])
            else:
                last = (ctx.status_code, ctx.text[:200])
        except requests.RequestException as err:
            last = err

        time.sleep(3)

    raise TimeoutError(f"could not log in as {email} within {ready_timeout}s (last={last})")


def mint_service_account_key(
    endpoint: str,
    bearer_token: str,
    *,
    name: str = SERVICE_ACCOUNT_NAME,
    role: str = SERVICE_ACCOUNT_ROLE,
) -> tuple[str, str]:
    """Create a service account, assign it a role, and return its ID and a fresh API key."""
    headers = {"Authorization": f"Bearer {bearer_token}"}

    sa = requests.post(f"{endpoint}/api/v1/service_accounts", json={"name": name}, headers=headers, timeout=10)
    assert sa.status_code == 201, sa.text
    sa_id = sa.json()["data"]["id"]

    roles = requests.get(f"{endpoint}/api/v1/roles", headers=headers, timeout=10)
    assert roles.status_code == 200, roles.text
    role_id = next(r["id"] for r in roles.json()["data"] if r["name"] == role)

    assign = requests.post(
        f"{endpoint}/api/v1/service_account_roles",
        json={"serviceAccountId": sa_id, "roleId": role_id},
        headers=headers,
        timeout=10,
    )
    assert assign.status_code == 201, assign.text

    key = requests.post(
        f"{endpoint}/api/v1/service_accounts/{sa_id}/keys",
        json={"name": "terraform-integration", "expiresAt": 0},
        headers=headers,
        timeout=10,
    )
    assert key.status_code == 201, key.text

    logger.info("minted service-account key for %s (role %s)", name, role)
    return sa_id, key.json()["data"]["key"]


def delete_service_account(instance: FuseScope) -> None:
    if not instance.service_account_id:
        return

    resp = requests.delete(
        f"{instance.endpoint}/api/v1/service_accounts/{instance.service_account_id}",
        headers={"Authorization": f"Bearer {instance.root_token}"},
        timeout=10,
    )
    logger.info("deleted service account %s -> %d", instance.service_account_id, resp.status_code)


@pytest.fixture(scope="session")
def fusescope(request: pytest.FixtureRequest, pytestconfig: pytest.Config) -> FuseScope:
    """A running FuseScope backend and an API key for Terraform.

    The suite does not start the backend: FUSESCOPE_ENDPOINT must point at one.
    FUSESCOPE_ACCESS_TOKEN, when set, is used as is; otherwise a service account
    is minted through the root user.
    """
    endpoint = os.environ.get(ENV_ENDPOINT, "").rstrip("/")
    if not endpoint:
        pytest.skip(f"{ENV_ENDPOINT} is not set; point it at a running FuseScope backend")

    def empty() -> FuseScope:
        return FuseScope(endpoint="", access_token="")

    def create() -> FuseScope:
        token = os.environ.get(ENV_ACCESS_TOKEN, "")
        if token:
            return FuseScope(endpoint=endpoint, access_token=token)

        root_token = _login_as_root(
            endpoint,
            os.environ.get(ENV_ROOT_EMAIL, DEFAULT_ROOT_EMAIL),
            os.environ.get(ENV_ROOT_PASSWORD, DEFAULT_ROOT_PASSWORD),
        )
        sa_id, access_token = mint_service_account_key(endpoint, root_token)
        return FuseScope(endpoint=endpoint, access_token=access_token, root_token=root_token, service_account_id=sa_id)

    def restore(cache: dict) -> FuseScope:
        return FuseScope(**cache)

    return reuse.wrap(request, pytestconfig, "fusescope", empty, create, delete_service_account, restore)
