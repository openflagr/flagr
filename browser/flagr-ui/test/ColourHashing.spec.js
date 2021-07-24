import helpers from '../src/helpers/helpers.js';

describe('Colour hashing helper', () => {
  test('should produce a valid css hsla string', () => {
    const randomString = Math.random().toString(36).substring(2, 15);
    expect(helpers.stringToColour(randomString)).toMatch(/hsla\((\d{1,3}), (\d{1,3}\%), (\d{1,3}\%), 1\)/);
  })
});
