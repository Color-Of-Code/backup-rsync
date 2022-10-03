
const report = async function (json) {
  const header = [
    '',
    'Line coverage',
    'Branch coverage'
  ];

  const metrics = [
    'Total',
    `${json.result.line} %`,
    `${json.result.branch} %`
  ];

  const message = `## SimpleCov Coverage
| | Line coverage | Branch coverage |
|---:|:---:|:---:|
| Total | ${json.result.line} % | ${json.result.branch} % |
`;

  return message;
}

module.exports = report;
