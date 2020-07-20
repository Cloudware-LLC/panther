const core = require('@actions/core');
const github = require('@actions/github');

const PR_TITLE_PREFIX = '[OSS Sync]';
const MASTER_BRANCH = 'v1.0.1-docs';

const main = async () => {
  try {
    const destRepo = core.getInput('destRepo');
    const ignoreLabel = core.getInput('ignoreLabel');
    const token = core.getInput('token');

    // Get the JSON webhook payload for the event that triggered the workflow
    const pullRequest = github.context.payload.pull_request;

    // If PR was closed, but it was not due to it being merged, then do nothing
    if (!pullRequest.merged) {
      core.debug('PR was closed without merging. Terminating...');
      core.setFailed('PR was not merged');
    }
    core.debug('PR was closed due to a merge. Looking for ignore labels...');

    // If PR has the "ignore" label, then the PR sync should not happen
    const shouldIgnore = pullRequest.labels.some(label => label.name === ignoreLabel);
    if (shouldIgnore) {
      core.debug('PR contained an ignore label. Terminating...');
      core.setFailed('PR was ignored by Author');
    }
    core.debug('An ignore label was not found. Starting sync process...');

    core.debug('Initializing octokit...');
    const octokit = github.getOctokit(token);
    core.debug('Octokit instance setup successfully');

    // https://developer.github.com/v3/git/refs/#create-a-reference
    core.debug('Creating a branch from the merge commit...');
    const prBranchName = pullRequest.head.ref;
    await octokit.request(`POST /repos/${destRepo}/git/refs`, {
      ref: `refs/heads/${prBranchName}`,
      sha: pullRequest.merge_commit_sha,
    });

    // https://developer.github.com/v3/pulls/#create-a-pull-request
    core.debug('Creating a pull request...');
    const { data: destPullRequest } = await octokit.request(`POST /repos/${destRepo}/pulls`, {
      title: `${PR_TITLE_PREFIX} ${pullRequest.title}`,
      body: pullRequest.body,
      maintainer_can_modify: true,
      head: prBranchName,
      base: MASTER_BRANCH,
      draft: false,
    });

    // https://developer.github.com/v3/issues/#update-an-issue
    core.debug('Setting assignees, labels & milestone...');
    core.debug(JSON.stringify(destPullRequest, null, 2));
    await octokit.request(`PATCH /repos/${destRepo}/issues/${destPullRequest.id}`, {
      assignees: pullRequest.assignees.map(assignee => assignee.login),
      labels: pullRequest.labels.map(label => label.name),
      milestone: pullRequest.milestone ? pullRequest.milestone.id : null,
    });

    // https://developer.github.com/v3/pulls/review_requests/#request-reviewers-for-a-pull-request
    core.debug('Setting reviewers...');
    await octokit.request(
      `POST /repos/${destRepo}/pulls/${destPullRequest.id}/requested_reviewers`,
      {
        reviewers: pullRequest.user.login,
      }
    );

    // Set the `url` output to the created PR's URL
    core.setOutput('url', destPullRequest.url);
  } catch (error) {
    core.setFailed(error);
  } finally {
    // noop
  }
};

main();
