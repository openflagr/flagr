function pluck(arr, prop) {
  return arr.map(el => el[prop])
}

function sum(arr) {
  return arr.reduce((acc, el) => acc + el, 0)
}
function debounce(fn, delay) {
  let timer = null;
  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delay);
  };
}

function handleErr(err) {
  const msg = err?.response?.data?.message || 'request error'
  this.$message.error(msg)
  if (err?.response?.status === 401) {
    const redirectURL = err?.response?.headers?.['www-authenticate']?.split('"')[1]
    if (redirectURL) window.location = redirectURL
  }
}

export default {
  pluck,
  sum,
  handleErr,
  debounce,
}
