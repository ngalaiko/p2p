function handleMessage(conn, msg) {
  console.log('received', msg)

  switch (msg.type) {
    case 'init':
      addPeer(msg.peer, true)
      selectPeer(msg.peer)

      var known = {}
      for (id in msg.peer.known_peers) {
        addPeer(msg.peer.known_peers[id], false)
        known['peer-'+id] = true
      }
      known['peer-'+msg.peer.id] = true

      document.querySelectorAll('.peer').forEach(e => {
        if (known[e.id]) {
          return
        }

        console.log('removing', e)
        e.parentNode.removeChild(e)
      })
      return
    case 'peer_added':
      addPeer(msg.peer, false)
      return
    case 'text_sent':
      addPeer(msg.message.to, false)
      addMessage(msg.message)
      return
    case 'text_received':
      addPeer(msg.message.from, false)
      addMessage(msg.message)
      return
    default:
      console.error('unknown message type', msg.type)
  }
}

function sendMessage(conn) {
  var msg = document.getElementById('message')

  if (msg.value === '') {
    return
  }

  let recipient = document.querySelector('.peer.active')

  var message = {
    type: 'text_sent',
    message: {
      text: msg.value,
      to: recipient.peer,
    },
  }

  conn.send(JSON.stringify(message))

  msg.value = ''
}

function addPeer(peer, isSelf) {
  if (document.getElementById('peer-'+peer.id) !== null) {
    return 
  }

  if (isSelf) {
    peer.name += ' (you)'
    peer.self = true
  }

  appendPeer(peer)
}

function appendPeer(peer) {
  var peerEntry = document.createElement('li')
  peerEntry.className = 'list-group-item d-flex justify-content-between align-items-center peer'
  peerEntry.id = 'peer-' + peer.id
  peerEntry.innerHTML = peer.name
  peerEntry.peer = peer
  peerEntry.onclick = function() {
    selectPeer(peer)
  }

  document.getElementById('peers').appendChild(peerEntry)
}

function selectPeer(peer) {
  selectPeerContact(peer)
  selectPeerChat(peer)
}

function getPeerChat(peer) {
  var chat = document.getElementById('peer-chat-'+peer.id)

  if (chat !== null) {
    return chat 
  }

  chat = document.createElement('div')
  chat.className = 'chat list-group list-group-flush flex-grow-1 collapse'
  chat.id = 'peer-chat-'+peer.id

  document.getElementById('messages').appendChild(chat)

  return chat
}

function selectPeerChat(peer) {
  var chat = getPeerChat(peer)

  document.querySelectorAll('.chat').forEach(e => {
    e.className += ' collapse'
  })

  chat.classList.remove('collapse')
}

function addMessage(msg) {
  var message = document.getElementById('message-'+msg.id)
  if (message !== null) {
    return 
  }

  var self
  document.querySelectorAll('.peer').forEach(e => {
    if (e.peer.self) {
      self = e.peer
    }
  })


  message = document.createElement('div')
  message.id = 'message-'+msg.id
  message.className = 'm-2'
  message.innerHTML = msg.text

  if (msg.from.id !== self.id) {
    message.className += ' text-left'
    getPeerChat(msg.from).appendChild(message)
    updateUnread(msg.from, 1)
  } else {
    message.className += ' text-right'
    getPeerChat(msg.to).appendChild(message)
  }
}

function selectPeerContact(peer) {
  document.querySelectorAll('.peer.active').forEach(p => {
    p.classList.remove('active')
  })

  document.getElementById('peer-'+peer.id).className += ' active'
  updateUnread(peer, -10000)
}

function updateUnread(peer, diff) {
  var peer = document.getElementById('peer-'+peer.id)
  var badge = document.getElementById('peer-unread-'+peer.id)

  if (badge === null) {
    var badge = document.createElement('li')
    badge.className = 'badge badge-primary badge-pill'
    badge.id = 'peer-unread-'+peer.id
    badge.innerHTML = 0
    peer.appendChild(badge)
  }

  badge.innerHTML = Number(badge.innerHTML) + diff

  if (peer.classList.contains('active')) {
    badge.innerHTML = ''
  }

  if (Number(badge.innerHTML) <= 0) {
    badge.innerHTML = ''
  }
}

function connectWs() {
    let wsAddr = 'ws://' + document.location.host + '/ws'
    let conn = new ReconnectingWebSocket(wsAddr);
    conn.onopen = function () {
      console.log('connected')
    }
    conn.onclose = function (evt) {
      console.log('connection closed')
    };
    conn.onmessage = function (evt) {
        var messages = evt.data.split('\n');
        for (var i = 0; i < messages.length; i++) {
            if (messages[i].length === 0) {
              continue
            }

            handleMessage(conn, JSON.parse(messages[i]))
        }
    };

    document.getElementById('send').onclick = function() {
      sendMessage(conn)
    }

    document.onkeypress = function(e) {
      if (e.key !== 'Enter') {
        return
      }
      sendMessage(conn)
    }
}

window.onload = connectWs
