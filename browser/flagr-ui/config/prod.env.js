module.exports = {
  NODE_ENV: '"production"',
  API_URL: '"/api/v1"',

  // ',' separated string
  // For example
  // FLAGR_UI_POSSIBLE_ENTITY_TYPES=report_int_id,account_int_id
  FLAGR_UI_POSSIBLE_ENTITY_TYPES: process.env.FLAGR_UI_POSSIBLE_ENTITY_TYPES
}
