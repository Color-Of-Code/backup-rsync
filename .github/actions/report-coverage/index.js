const core = require('@actions/core');
const path = require('path');
const report = require('./src/report');
const comment = require('./src/comment');

async function run() {
  try {
    const jsonPath = path.resolve(process.env.GITHUB_WORKSPACE, 'coverage/.last_run.json');
    core.debug(`jsonPath: ${jsonPath}`);

    const json = require(jsonPath);
    const message = await report(json);
    core.debug(`message: ${message}`);

    await comment(message);
    core.debug(`added comment`);

  } catch (error) {
    core.setFailed(error.message);
  }
}

run();
