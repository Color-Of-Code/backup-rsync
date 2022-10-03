const report = require('./src/report')
const path = require('path')

test('json with group metrics', async () => {
  const json = require(path.resolve('./', 'examples/result_good.json'));
  const expected = `## SimpleCov Coverage
  |  | Line coverage | Branch coverage |
| Total | 93.33 | 100 |
`;
  expect(await report(json)).toEqual(expected);
})
