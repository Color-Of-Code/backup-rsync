const core = require('@actions/core');
const path = require('path')
const report = require('./src/report');

async function run() {
  try {
    const json = require(path.resolve(process.env.GITHUB_WORKSPACE, 'coverage/.last_run.json'));
    report(json);

  } catch (error) {
    core.setFailed(error.message);
  }
}

run();
