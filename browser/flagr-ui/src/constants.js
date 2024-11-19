const API_URL = process.env.VUE_APP_API_URL
const SSO_URL = process.env.VUE_APP_SSO_API_URL
const FLAGR_UI_POSSIBLE_ENTITY_TYPES = process.env.VUE_APP_FLAGR_UI_POSSIBLE_ENTITY_TYPES
const API_URLS = {
  USER_DETAILS: "api/v1/user/details",
  REFRESH_TOKEN: "api/v1/auth/authenticate/refresh",
}

const ENVS = {
  PROD: 'PROD',
  STAGE: 'STAGE',
  DEV: 'DEV'
}

const ENVURLS = {
  PROD: {
    VUE_APP_API_URL : 'http://flagr-new.allen-live.in/api/v1',
    VUE_APP_SSO_API_URL : 'https://api.allen-live.in/internal-bff/',
  },
  STAGE: {
    VUE_APP_API_URL : 'https://flagr-new.allen-stage.in/api/v1',
    VUE_APP_SSO_API_URL : 'https://bff.allen-stage.in/internal-bff/',
  },
  DEV: {
    VUE_APP_API_URL : 'http://flagr-new.allen-demo.in/api/v1',
    VUE_APP_SSO_API_URL : 'https://bff-dev.allen-demo.in/internal-bff/',
  }
}

export default {
  API_URL,
  SSO_URL,
  FLAGR_UI_POSSIBLE_ENTITY_TYPES,
  API_URLS,
  ENVS,
  ENVURLS
}
