const core = require('@actions/core');
const github = require('@actions/github');

const PR_TITLE_PREFIX = '[OSS Sync]';
const MASTER_BRANCH = 'master';

const main = async () => {
  try {
    const destRepo = core.getInput('destRepo');
    const ignoreLabel = core.getInput('ignoreLabel');
    const token = core.getInput('token');

    // Get the JSON webhook payload for the event that triggered the workflow
    const { action, pull_request: pullRequest } = github.context.payload;

    // If PR is still open, then do nothing
    if (action !== 'closed') {
      return;
    }

    // If PR was closed, but it was not due to it being merged, then do nothing
    if (!pullRequest.merged) {
      return;
    }

    // If PR has the "ignore" label, then the PR sync should not happen
    const isBackport = pullRequest.labels.some(label => label.name === ignoreLabel);
    if (isBackport) {
      return;
    }

    // const octokit = new Octokit({ auth: token });
    const octokit = github.getOctokit(token);
    const destPullRequest = await octokit(`POST /repos/${destRepo}/pulls`, {
      title: `${PR_TITLE_PREFIX} ${pullRequest.title}`,
      body: pullRequest.body,
      maintainer_can_modify: true,
      head: pullRequest.head.label,
      base: MASTER_BRANCH,
      draft: false,
    });

    // Clone the existing labels
    await octokit(
      `POST /repos/${destRepo}/pulls``/repos/${destRepo}/issues/${destPullRequest.id}/labels`,
      {
        labels: pullRequest.labels.map(label => label.name),
      }
    );

    // Clone the existing labels
    await octokit(
      `POST /repos/${destRepo}/pulls``/repos/${destRepo}/issues/${destPullRequest.id}`,
      {
        labels: pullRequest.labels.map(label => label.name),
        milestone: pullRequest.milestone.id,
      }
    );

    // Set the `url` output to the created PR's URL
    core.setOutput('url', destPullRequest.url);
  } catch (error) {
    core.setFailed(error.message);
  } finally {
    // noop
  }
};

main();
