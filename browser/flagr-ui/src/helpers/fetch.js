function getJson (url, options) {
  return fetch(url, options).then(res => res.json())
}

function postJson (url, data) {
  const request = new Request(url, {
    method: 'post',
    body: JSON.stringify(data),
    headers: new Headers({
      'content-type': 'application/json'
    })
  })

  return fetch(request).then(res => res.json())
}

export default {
  getJson,
  postJson
}
