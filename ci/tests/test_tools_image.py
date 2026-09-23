"""Exercise registry failures without contacting a registry or building images.

Run with: python3 -B -m unittest discover -s ci/tests -v
"""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
MISSING = "manifest for example:tag not found: manifest unknown: manifest unknown"
DOCKER = """#!/usr/bin/env python3
import json, os, sys
from pathlib import Path
log = Path(os.environ['MOCK_LOG'])
calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
args = sys.argv[1:]
with log.open('a') as output:
    output.write(json.dumps(args) + '\\n')
if args[0] == 'pull':
    cache = ':ci-tools-cache-' in args[1]
    errors = json.loads(os.environ['MOCK_CACHE_ERRORS' if cache else 'MOCK_PRIMARY_ERRORS'])
    count = sum(call == args for call in calls)
    error = errors[min(count, len(errors) - 1)]
    if error:
        print(error.replace('{reference}', args[1]), file=sys.stderr)
        sys.exit(1)
if args[0] == 'login':
    assert sys.stdin.read() == os.environ['DOCKER_PASSWORD']
if args[0] == os.environ.get('MOCK_FAIL_ACTION'):
    sys.exit(1)
"""


class ToolsImageTest(unittest.TestCase):
    def run_case(self, *, primary=(None,), cache=(None,), env=None, runner=False):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            docker = directory / "docker"
            docker.write_text(DOCKER)
            docker.chmod(0o755)
            # Bound retries are checked by their pull counts, without real sleeps.
            sleep = directory / "sleep"
            sleep.write_text("#!/bin/sh\nexit 0\n")
            sleep.chmod(0o755)
            environment = os.environ.copy()
            for key in list(environment):
                if key.startswith(("CI_", "TRAVIS_", "DOCKER_", "PLUGINS_")):
                    environment.pop(key)
            environment.update({
                "PATH": f"{directory}:{environment['PATH']}",
                "MOCK_LOG": str(directory / "calls"),
                "MOCK_PRIMARY_ERRORS": json.dumps(primary),
                "MOCK_CACHE_ERRORS": json.dumps(cache),
                "CI_ARCH": "amd64",
                "CI_CACHE_ROOT": str(directory / "cache"),
                "CI_DOCKER_SOCKET": str(docker),
            })
            environment.update(env or {})
            command = ["bash", "ci/run.sh"] if runner else ["bash", "ci/tools-image.sh", "ensure"]
            result = subprocess.run(command, cwd=ROOT, env=environment, text=True,
                                    stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
            log = directory / "calls"
            calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
            return result, calls

    def assert_actions(self, result, calls, expected, success=True):
        self.assertEqual(result.returncode == 0, success, result.stdout)
        self.assertEqual([call[0] for call in calls], expected, result.stdout)

    def test_existing_image_only_pulls(self):
        result, calls = self.run_case(env={"CI_TOOLS_PUBLISH": "true"})
        self.assert_actions(result, calls, ["pull"])
        self.assertIn("TOOLS_IMAGE_CACHE=hit", result.stdout)

    def test_missing_image_reuses_previous_layer_cache(self):
        result, calls = self.run_case(primary=(MISSING,))
        self.assert_actions(result, calls, ["pull", "pull", "build"])
        self.assertIn("TOOLS_IMAGE_CACHE=miss", result.stdout)
        build = calls[-1]
        self.assertEqual(build[build.index("--cache-from") + 1], calls[1][1])
        self.assertEqual(build[build.index("--tag") + 1], calls[0][1])

    def test_missing_previous_cache_still_builds(self):
        result, calls = self.run_case(primary=(MISSING,), cache=(MISSING,))
        self.assert_actions(result, calls, ["pull", "pull", "build"])

    def test_containerd_missing_reference_builds(self):
        error = ('no matching manifest for linux/arm64/v8 in the manifest list entries: '
                 'failed to resolve reference "{reference}": {reference}: not found')
        result, calls = self.run_case(primary=(error,))
        self.assert_actions(result, calls, ["pull", "pull", "build"])

    def test_publisher_logs_in_and_pushes_only_after_successful_build(self):
        result, calls = self.run_case(primary=(MISSING,), env=self.publisher_env())
        self.assert_actions(result, calls, ["pull", "pull", "build", "login", "push", "tag", "push"])
        self.assertEqual(calls[4][1], calls[0][1])
        self.assertEqual(calls[5][1:], [calls[0][1], calls[1][1]])
        self.assertEqual(calls[6][1], calls[1][1])

    def test_missing_credentials_only_builds_locally(self):
        for credentials in ({}, {"DOCKER_USERNAME": "test-user"}, {"DOCKER_PASSWORD": "test-password"}):
            with self.subTest(credentials=list(credentials)):
                result, calls = self.run_case(primary=(MISSING,), env={"CI_TOOLS_PUBLISH": "true", **credentials})
                self.assert_actions(result, calls, ["pull", "pull", "build"])

    def test_transient_failure_retries_then_uses_image(self):
        result, calls = self.run_case(primary=("TLS handshake timeout", None))
        self.assert_actions(result, calls, ["pull", "pull"])
        self.assertIn("TOOLS_IMAGE_CACHE=hit", result.stdout)

    def test_registry_failures_never_trigger_build(self):
        for error in ("unauthorized: authentication required", "denied: requested access denied",
                      "toomanyrequests: pull rate limit", "TLS handshake timeout",
                      "proxy not found", "no matching manifest for linux/amd64"):
            with self.subTest(error=error):
                result, calls = self.run_case(primary=(error,))
                self.assert_actions(result, calls, ["pull"] * 3, success=False)

    def test_previous_cache_outage_is_not_treated_as_missing(self):
        result, calls = self.run_case(primary=(MISSING,), cache=("TLS handshake timeout",))
        self.assert_actions(result, calls, ["pull"] * 4, success=False)

    def test_build_or_publish_failure_stops_pipeline(self):
        actions = ["pull", "pull", "build", "login", "push", "tag", "push"]
        for failure in ("build", "login", "push", "tag"):
            with self.subTest(failure=failure):
                result, calls = self.run_case(primary=(MISSING,), env={**self.publisher_env(), "MOCK_FAIL_ACTION": failure})
                self.assert_actions(result, calls, actions[:actions.index(failure) + 1], success=False)

    def test_runner_designates_only_public_non_pr_job_to_publish(self):
        for visibility, request, publish in (("public", "false", True), ("private", "false", False),
                                             ("public", "123", False), ("public", "", False)):
            with self.subTest(visibility=visibility, request=request):
                environment = self.publisher_env()
                environment.pop("CI_TOOLS_PUBLISH")
                environment.update({"PLUGINS_VISIBILITY": visibility, "TRAVIS_PULL_REQUEST": request})
                result, calls = self.run_case(primary=(MISSING,), env=environment, runner=True)
                expected = ["pull", "pull", "build"]
                if publish:
                    expected += ["login", "push", "tag", "push"]
                self.assert_actions(result, calls, expected + ["run"])

    def test_runner_preserves_local_image_override(self):
        result, calls = self.run_case(env={"CI_TOOLS_LOCAL": "true"}, runner=True)
        self.assert_actions(result, calls, ["run"])

    def test_runner_stops_on_registry_failure(self):
        result, calls = self.run_case(primary=("unauthorized",), runner=True)
        self.assert_actions(result, calls, ["pull"] * 3, success=False)

    @staticmethod
    def publisher_env():
        return {"CI_TOOLS_PUBLISH": "true", "DOCKER_USERNAME": "test-user", "DOCKER_PASSWORD": "test-password"}


if __name__ == "__main__":
    unittest.main()
