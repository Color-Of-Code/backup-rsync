const core = require('@actions/core');
const github = require('@actions/github');

const comment = async function (message) {
  const pullRequestId = github.context.issue.number;
  if (pullRequestId) {
    const octokit = new github.getOctokit(process.env.GITHUB_TOKEN);
    const response = await octokit.rest.issues.createComment({
      owner: github.context.repo.owner,
      repo: github.context.repo.repo,
      issue_number: pullRequestId,
      body: message
    });
    core.debug(`created comment URL: ${response.data.html_url}`)
  }
}

module.exports = comment;
