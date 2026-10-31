import shutil
from collections.abc import Callable
from pathlib import Path

import pytest

from fixtures.fusescope import FuseScope
from fixtures.tool import EXAMPLES, Tool

RESOURCE_FILES = sorted((EXAMPLES / "resources").glob("fusescope_*/*.tf"))
RECIPES = sorted(p for p in (EXAMPLES / "recipes").iterdir() if p.is_dir())

# Example files to skip, keyed by "<resource>/<file>" -> reason. The integration
# run surfaced provider/API issues for these.
SKIPPED = {
    "fusescope_planned_maintenance/resource.tf": "alert_ids reference non-existent rules (API 500)",
    "fusescope_service_account_role/resource.tf": "needs a managed role ID (var.viewer_role_id); FuseScope has no custom roles",
    "fusescope_user_role/resource.tf": "needs a managed role ID (var.viewer_role_id); FuseScope has no custom roles",
}

# Values for the examples' input variables. Secrets are fake and URLs point at
# example.com, so a firing rule can never reach a real Slack, PagerDuty, etc.
VARIABLES = {
    "slack_webhook_url": "https://example.com/slack",
    "msteams_webhook_url": "https://example.com/msteams",
    "googlechat_webhook_url": "https://example.com/googlechat",
    "pagerduty_routing_key": "fake-pagerduty-routing-key",
    "opsgenie_api_key": "fake-opsgenie-api-key",
    "service_name": "terraform-integration",
}


@pytest.mark.parametrize(
    "tf_file",
    [pytest.param(tf_file, id=(rel := f"{tf_file.parent.name}/{tf_file.name}"), marks=[pytest.mark.skip(reason=SKIPPED[rel])] if rel in SKIPPED else []) for tf_file in RESOURCE_FILES],
)
def test_resource_file_crud(tf_file: Path, workspace: Callable[[Path], Path], tool_config: Path, fusescope: FuseScope, tool_bin: str):
    tool = Tool(workspace(tf_file), tool_config, fusescope, tool_bin, VARIABLES)

    tool.apply()

    try:
        assert tool.plan_exit_code() == 0, "drift detected after apply"
    finally:
        tool.destroy()


@pytest.mark.parametrize("recipe", RECIPES, ids=[r.name for r in RECIPES])
def test_recipe_crud(recipe: Path, tmp_path: Path, tool_config: Path, fusescope: FuseScope, tool_bin: str):
    for tf_file in recipe.glob("*.tf"):
        shutil.copy(tf_file, tmp_path / tf_file.name)

    tool = Tool(tmp_path, tool_config, fusescope, tool_bin, VARIABLES)

    tool.apply()

    try:
        assert tool.plan_exit_code() == 0, "drift detected after apply"
    finally:
        tool.destroy()
