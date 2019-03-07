function main() {
  var login = document.getElementById('login')

  let token = getCookie("auth")

  if (token === undefined) {
    return
  }

  let claims = parseJwt(token)

  login.innerHTML = 'Continue as ' + claims.Peer.name
}

function loginClick(redirect) {
  if (redirect === "") {
    return dispatch
  }
  return function() {
    // todo: implement
    console.log('redirect to', redirect)
  }
}

function dispatch() {
  // todo: implement
  console.log('start dispatching')
}

function parseJwt (token) {
  let base64Url = token.split('.')[1]
  let base64 = base64Url.replace('-', '+').replace('_', '/')
  return JSON.parse(window.atob(base64))
}

function getCookie(name) {
  let value = "; " + document.cookie
  let parts = value.split("; " + name + "=")
  if (parts.length == 2) return parts.pop().split(";").shift()
}

window.onload = main
