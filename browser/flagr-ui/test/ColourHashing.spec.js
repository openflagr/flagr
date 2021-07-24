import helpers from '@/helpers/helpers.js';
import { expect } from 'chai';

describe('Colour hashing helper', () => {
  it('should produce a valid css hsla string', () => {
    const randomString = Math.random().toString(36).substring(2, 15);
    expect(helpers.stringToColour(randomString)).to.match(/hsla\((\d{1,3}), (\d{1,3}\%), (\d{1,3}\%), 1\)/);
  })
});
