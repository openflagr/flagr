const ENV = process.env.NODE_ENV
const API_URL = process.env.API_URL
const DEV = ENV === 'development'

export default {
  ENV,
  DEV,
  API_URL
}
