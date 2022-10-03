const table = rows =>
  rows.map(x => '| ' + x.join(' | ') + ' |').join("\n");

const report = async function (json) {
  const groups = json.groups || [];

  const header = [
    '',
    'Line coverage',
    'Branch coverage'
  ];

  const metrics = [
    'Total',
    json.result.line,
    json.result.branch
  ];

  const tableText = table([header, metrics]);
  const message = `## SimpleCov Coverage
  ${tableText}
`;

  return message;
}

module.exports = report;
