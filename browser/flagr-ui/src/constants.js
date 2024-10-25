const API_URL = process.env.VUE_APP_API_URL
const SSO_URL = process.env.VUE_APP_SSO_API_URL
const FLAGR_UI_POSSIBLE_ENTITY_TYPES = process.env.VUE_APP_FLAGR_UI_POSSIBLE_ENTITY_TYPES
const API_URLS = {
  USER_DETAILS: "api/v1/user/details"
}
export const MODES = {
  ABMode: "ABMode",
  LatchMode: "LatchMode",
}

export const ABModeConstants = {
  flag: 'Flag',
  segment: 'Segment',
  constraint: 'Constraint'
}

export const LatchModeConstants = {
  flag: 'Latch',
  segment: 'Cohort',
  constraint: 'Lever'
}
export default {
  API_URL,
  SSO_URL,
  FLAGR_UI_POSSIBLE_ENTITY_TYPES,
  API_URLS
}
