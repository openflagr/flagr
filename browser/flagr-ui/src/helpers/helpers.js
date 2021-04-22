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
  this.$message.error(msg)
  if (get(err, 'response.status') === 401) {
    let redirectURL = err.response.headers['www-authenticate'].split(`"`)[1]
    window.location = redirectURL
    return
  }
}

function getHue(str) {
  var hash = 0;
  if (str.length === 0) return hash;
  for (var i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
    hash = hash & hash;
  }
  return hash % 360;
}

function stringToColour(str, lightness = 90) {
  let hue = getHue(str);

  return `hsla(${hue}, 100%, ${lightness}%, 1)`;
}

export default {
  indexBy,
  pluck,
  sum,
  get,
  handleErr,
  stringToColour
};
