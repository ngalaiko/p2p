function handleMessage(conn, msg) {
  console.log("received", msg)

  switch (msg.type) {
    case "init":
      updatePeers(msg.self, true)

      for (id in msg.self.known_peers) {
        updatePeers(msg.self.known_peers[id], false)
      }
      return
    case "peers_update":
      updatePeers(msg.self, true)

      for (id in msg.self.known_peers) {
        updatePeers(msg.self.known_peers[id], false)
      }
      return
    default:
      console.error("unknown message type", msg.type)
  }
}

function updatePeers(peer, isSelf) {
  if (document.getElementById('peer-'+peer.id) !== null) {
    return 
  }

  if (isSelf) {
    peer.name += " (me)"
  }

  appendPeer(peer)
}

function appendPeer(peer) {
  var peerEntry = document.createElement("li")
  peerEntry.className = "list-group-item d-flex justify-content-between align-items-center peer"
  peerEntry.id = "peer-" + peer.id
  peerEntry.innerHTML = peer.name
  peerEntry.onclick = selectPeer(peer)    

  document.getElementById("peers").appendChild(peerEntry)
}

function selectPeer(peer) {
  return function() {
    document.querySelectorAll(".peer.active").forEach(p => {
      p.classList.remove("active")
    })

    document.getElementById("peer-"+peer.id).className += " active"
  }
}

function updateUnread(peer, diff) {
  var peer = document.getElementById("peer-"+peer.id)
  var badge = document.getElementById("peer-unread-"+peer.id)

  if (badge === null) {
    var badge = document.createElement("li")
    badge.className = "badge badge-primary badge-pill"
    badge.id = "peer-unread-"+peer.id
    badge.innerHTML = 0
    peer.appendChild(badge)
  }

  badge.innerHTML = Number(badge.innerHTML) + diff

  if (Number(badge.innerHTML) <= 0) {
    badge.innerHTML = ""
  }
}

function connectWs() {
    let wsAddr = "ws://" + document.location.host + "/ws"
    let conn = new WebSocket(wsAddr);
    conn.onclose = function (evt) {
      console.log("connection closed")
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
}

window.onload = connectWs

//window.onload = function () {
    //let msg = document.getElementById("msg");
    //let log = document.getElementById("log");

    //let peersDiv = document.getElementById("peers")
    //var peersList = {};

    //function appendLog(item) {
        //var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        //log.appendChild(item);
        //if (doScroll) {
            //log.scrollTop = log.scrollHeight - log.clientHeight;
        //}
    //}
    //function updatePeers(newPeers) {
      //for (let id in newPeers) {
        //if (peersList[id] !== undefined) {
          //console.log("known peer", id)
          //return
        //}

        //console.log("adding new peer", id)

        //peersList[id] = newPeers[id]

        //var item = document.createElement("div")
        //item.innerHTML = id
        //item.peer = peers[id]

        //console.log(item)

        //var doScroll = peersDiv.scrollTop > peersDiv.scrollHeight - peersDiv.clientHeight - 1;
        //peersDiv.appendChild(item)
        //if (doScroll) {
          //peersDiv.scrollTop = peersDiv.scrollHeight - peersDiv.clientHeight;
        //}
      //}
      //return
    //}
    //if (!window["WebSocket"]) {
      //var item = document.createElement("div");
      //item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
      //appendLog(item);
      //return
    //}
    //document.getElementById("form").onsubmit = function () {
        //if (!conn) {
            //return false;
        //}
        //if (!msg.value) {
            //return false;
        //}
        //conn.send(msg.value);
        //msg.value = "";
        //return false;
    //};

    //function processMessage(msg) {
      //var item = document.createElement("div");

      //console.log("received", msg)

      //switch (msg.type) {
        //case "init":
          //updatePeers(msg.self.known_peers)
        //case "peers_update":
          //updatePeers(msg.self.known_peers)
      //}
    //}

//};

