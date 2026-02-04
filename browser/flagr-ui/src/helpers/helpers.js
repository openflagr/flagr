import { ElMessage } from 'element-plus'

function indexBy (arr, prop) {
  return arr.reduce((acc, el) => {
    acc[el[prop]] = el
    return acc
  }, {})
}

function pluck (arr, prop) {
  return arr.map(el => el[prop])
}

function sum (arr) {
  return arr.reduce((acc, el) => {
    acc += el
    return acc
  }, 0)
}

function get (obj, path, def) {
  const fullPath = path
    .replace(/\[/g, '.')
    .replace(/]/g, '')
    .split('.')
    .filter(Boolean)

  return fullPath.every(everyFunc) ? obj : def

  function everyFunc (step) {
    return !(step && (obj = obj[step]) === undefined)
  }
}

function handleErr (err) {
  let msg = get(err, 'response.data.message', 'request error')
  ElMessage.error(msg)
  if (get(err, 'response.status') === 401) {
    let redirectURL = err.response.headers['www-authenticate'].split(`"`)[1]
    window.location = redirectURL
  }
}

export default {
  indexBy,
  pluck,
  sum,
  get,
  handleErr
}
