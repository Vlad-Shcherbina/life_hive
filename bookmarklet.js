// to minify: http://jscompress.com/

(function () {
  function interact() {
    var players_ = app.playerManager.players;
    var players = [];
    for (var i = 0; i < players_.length; i++) {
      players.push({
        Id: players_[i].id,
        Name: players_[i].name,
        Color: players_[i].color,
        Online: players_[i].online
      });
    }

    var my_player = app.playerManager.getLocalPlayer();
    var grid = app.game.grid;

    cells = []
    for (var i = 0; i < grid.height; i++) {
      var row = []
      for (var j = 0; j < grid.width; j++) {
        var cell = grid.getCell(j, i);
        if (cell.alive) {
          row.push(cell.playerId);
        } else {
          row.push(0);
        }
      }
      cells.push(row);
    }

    var data = {
      MyId: my_player.id,
      Cells: cells,
      Players: players,
    };
    console.log('to send', data);

    var xmlhttp = new XMLHttpRequest();
    var url = "http://localhost:8000/bot";
    xmlhttp.open("POST", url, true);
    xmlhttp.setRequestHeader("Content-type","application/json");
    xmlhttp.send(JSON.stringify(data) + "\n");
    console.log('sent', xmlhttp);
    xmlhttp.onreadystatechange = function() {
      if (xmlhttp.readyState != 4) return;
      if (xmlhttp.status != 200) {
        console.log('error', xmlhttp.responseText);
        return;
      }
      var data = JSON.parse(xmlhttp.responseText);
      console.log('received', data);
    };
  }

  setInterval(interact, 1000);
})();
