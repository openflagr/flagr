function getJson (url, options) {
  return fetch(url, options).then(res => res.json())
}

function requestJson (method, url, data) {
  const request = new Request(url, {
    method,
    body: JSON.stringify(data),
    headers: new Headers({
      'content-type': 'application/json'
    })
  })

  return fetch(request).then(res => res.json())
}

function postJson (url, data) {
  return requestJson('post', url, data)
}

function putJson (url, data) {
  return requestJson('put', url, data)
}

export default {
  getJson,
  postJson,
  putJson
}
