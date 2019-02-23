window.onload = function () {
    var conn;
    let msg = document.getElementById("msg");
    let log = document.getElementById("log");

    let peersDiv = document.getElementById("peers")
    var peersList = {};

    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }
    function updatePeers(newPeers) {
      for (let id in newPeers) {
        if (peersList[id] !== undefined) {
          console.log("known peer", id)
          return
        }

        console.log("adding new peer", id)

        peersList[id] = newPeers[id]

        var item = document.createElement("div")
        item.innerHTML = id
        item.peer = peers[id]

        console.log(item)

        var doScroll = peersDiv.scrollTop > peersDiv.scrollHeight - peersDiv.clientHeight - 1;
        peersDiv.appendChild(item)
        if (doScroll) {
          peersDiv.scrollTop = peersDiv.scrollHeight - peersDiv.clientHeight;
        }
      }
      return
    }
    if (!window["WebSocket"]) {
      var item = document.createElement("div");
      item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
      appendLog(item);
      return
    }
    document.getElementById("form").onsubmit = function () {
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }
        conn.send(msg.value);
        msg.value = "";
        return false;
    };

    function processMessage(msg) {
      var item = document.createElement("div");

      console.log("received", msg)

      switch (msg.type) {
        case "init":
          updatePeers(msg.self.known_peers)
        case "peers_update":
          updatePeers(msg.self.known_peers)
      }
    }

    let wsAddr = "ws://" + document.location.host + "/ws"
    conn = new WebSocket(wsAddr);
    conn.onclose = function (evt) {
        var item = document.createElement("div");
        item.innerHTML = "<b>Connection closed.</b>";
        appendLog(item);
    };
    conn.onmessage = function (evt) {
        var messages = evt.data.split('\n');
        for (var i = 0; i < messages.length; i++) {
            if (messages[i].length === 0) {
              continue
            }

            processMessage(JSON.parse(messages[i]))
        }
    };
};

