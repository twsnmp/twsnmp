'use strict';

let myFont;
let selectNode = "";
let mapConf;
let notifyConf;
let nodes = {};
let lines = {};
let backimg;

const status = {
  High: 0,
  Low: 0,
  Warn: 0,
  Normal: 0,
  Repair: 0,
  Unkown: 0
};

function preload() {
  myFont = loadFont('./webfonts/fa-solid-900.ttf');
}

function setup() {
  var canvas = createCanvas(2000, 2000);
  canvas.parent('mapDiv');
  noLoop();
}

function draw() {
  background(250);
  if(backimg){
    image(backimg,0,0);
  }
  for (let k in lines) {
    if (!nodes[lines[k].NodeID1] || !nodes[lines[k].NodeID2]) {
      continue;
    }
    const x1 = nodes[lines[k].NodeID1].X;
    const x2 = nodes[lines[k].NodeID2].X;
    const y1 = nodes[lines[k].NodeID1].Y + 6;
    const y2 = nodes[lines[k].NodeID2].Y + 6;
    const xm = (x1 + x2) / 2;
    const ym = (y1 + y2) / 2;
    push();
    strokeWeight(2);
    stroke(getStateColor(lines[k].State1));
    line(x1, y1, xm, ym);
    stroke(getStateColor(lines[k].State2));
    line(xm, ym, x2, y2);
    pop();
  }
  for (let k in nodes) {
    const icon = getIcon(nodes[k].Icon);
    push();
    translate(nodes[k].X, nodes[k].Y);
    if (selectNode == nodes[k].ID) {
      fill("aliceblue");
      stroke(getStateColor(nodes[k].State));
      rect(-24, -24, 48, 48);
    } else {
      fill(250);
      stroke(250);
      rect(-18, -18, 36, 36);
    }
    textFont(myFont);
    textSize(32);
    textAlign(CENTER, CENTER);
    fill(0);
    text(icon, 0, 0);
    fill(getStateColor(nodes[k].State));
    text(icon, -1, -1);
    textFont("Arial");
    textSize(12);
    fill(0);
    text(nodes[k].Name, 0, 32);
    pop();
  }
}

function setSelectNode() {
  for (let k in nodes) {
    if (nodes[k].X + 32 > mouseX &&
      nodes[k].X - 32 < mouseX &&
      nodes[k].Y + 32 > mouseY &&
      nodes[k].Y - 32 < mouseY
    ) {
      selectNode = nodes[k].ID;
      return;
    }
  }
  selectNode = "";
  return;
}

let lastMouseX;
let lastMouseY;
let draggedNode = "";

function mouseDragged() {
  if (winMouseX < 200 ||
    winMouseY < 32 ||
    winMouseY > windowHeight * 0.75) {
    return true;
  }
  if (nodes[selectNode] && lastMouseX) {
    nodes[selectNode].X += mouseX - lastMouseX;
    nodes[selectNode].Y += mouseY - lastMouseY;
    if (nodes[selectNode].X < 16) {
      nodes[selectNode].X = 16;
    }
    if (nodes[selectNode].Y < 16) {
      nodes[selectNode].Y = 16;
    }
    draggedNode = selectNode;
    redraw();
  }
  lastMouseX = mouseX;
  lastMouseY = mouseY;
  return true;
}

let ctxMenu;

function mousePressed() {
  // クリックした位置がマップ以外は、処理しない。
  if (winMouseX < 200 ||
    winMouseY < 32 ||
    winMouseY > windowHeight * 0.75) {
    return true;
  }
  if (ctxMenu) {
    return true;
  }

  const selectNodeBack = selectNode;
  setSelectNode();
  if (keyIsDown(SHIFT) &&
    selectNodeBack != "" &&
    selectNode != "" &&
    selectNodeBack != selectNode) {
    createEditLinePane(selectNodeBack, selectNode);
    selectNode = "";
    return true;
  }
  if (selectNodeBack != selectNode) {
    updateNodeList();
  }
  if (mouseButton === RIGHT) {
    let div;
    if (nodes[selectNode]) {
      div = `
      <nav class="nav-group">
        <span class="nav-group-item showNodeInfo">
          <i class="fas fa-info-circle"></i>    
         ノード情報
        </span>
        <span class="nav-group-item showPolling">
          <i class="fas fa-info-circle"></i>    
         ポーリング
        </span>
        <span class="nav-group-item showNodeLog">
          <i class="fas fa-info-circle"></i>    
         ログ
        </span>
        <span class="nav-group-item pollNow">
          <i class="fas fa-info-circle"></i>    
         再確認
        </span>
        <span class="nav-group-item showMIB">
          <i class="fas fa-info-circle"></i>    
         MIBブラウザ
        </span>
        <span class="nav-group-item editNode">
          <i class="fas fa-cog"></i>
          編集
        </span>
        <span class="nav-group-item dupNode">
          <i class="fas fa-copy"></i>
          複製
        </span>
        <span class="nav-group-item deleteNode">
          <i class="fas fa-trash-alt"></i>
          削除...
        </span>
      </nav>
      `;
    } else {
      div = `
      <nav class="nav-group">
        <span class="nav-group-item startDiscover">
          <i class="fas fa-search"></i>
          自動発見
        </span>
        <span class="nav-group-item addNode">
          <i class="fas fa-plus-circle"></i>
          新規ノード
        </span>
        <span class="nav-group-item mapConf">
          <i class="fas fa-cog"></i>
          マップ設定
        </span>
        <span class="nav-group-item notifyConf">
          <i class="fas fa-mail-bulk"></i>
          通知設定
        </span>
        <span class="nav-group-item actionPane">
          <i class="fas fa-trash-restore"></i>
          削除操作...
        </span>
        <span class="nav-group-item checkAllPoll">
        <i class="fas fa-check-square"></i>
          全て再確認...
        </span>
        <span class="nav-group-item showPollingList">
          <i class="fas fa-exchange-alt"></i>
          ポーリングリスト
        </span>
        <span class="nav-group-item logDisp">
          <i class="fas fa-clipboard-list"></i>
          ログ表示
        </span>
      </nav>
      `;
    }
    ctxMenu = createDiv(div);
    ctxMenu.id("ctxMenu");
    ctxMenu.position(winMouseX, winMouseY + 10);
    $("#ctxMenu span.deleteNode").on("click", () => {
      deleteNode();
    });
    $("#ctxMenu span.dupNode").on("click", () => {
      dupNode();
    });
    $("#ctxMenu span.showNodeInfo").on("click", () => {
      if (selectNode != "") {
        astilectron.sendMessage({ name: "showNodeInfo", payload: selectNode }, function (message) {
        });
      }
    });
    $("#ctxMenu span.showPolling").on("click", () => {
      if (selectNode != "") {
        astilectron.sendMessage({ name: "showPolling", payload: selectNode }, function (message) {
        });
      }
    });
    $("#ctxMenu span.showNodeLog").on("click", () => {
      if (selectNode != "") {
        astilectron.sendMessage({ name: "showNodeLog", payload: selectNode }, function (message) {
        });
      }
    });
    $("#ctxMenu span.pollNow").on("click", () => {
      if (selectNode != "") {
        astilectron.sendMessage({ name: "pollNow", payload: selectNode }, function (message) {
          nodes[selectNode].State = "unkown";
          redraw();
        });
      }
    });
    $("#ctxMenu span.showMIB").on("click", () => {
      if (selectNode != "" ) {
        astilectron.sendMessage({ name: "showMIB", payload: selectNode }, function (message) {
        });
      }
    });
    $("#ctxMenu span.editNode").on("click", () => {
      if (selectNode != "") {
        createEditNodePane(lastMouseX, lastMouseY, selectNode);
      }
    });
    $("#ctxMenu span.startDiscover").on("click", () => {
      createStartDiscoverPane(lastMouseX, lastMouseY);
    });
    $("#ctxMenu span.addNode").on("click", () => {
      createEditNodePane(lastMouseX, lastMouseY, "");
    });
    $("#ctxMenu span.mapConf").on("click", () => {
      createMapConfPane();
    });
    $("#ctxMenu span.notifyConf").on("click", () => {
      createNotifyConfPane();
    });
    $("#ctxMenu span.actionPane").on("click", () => {
      createActionPane();
    });
    $("#ctxMenu span.logDisp").on("click", () => {
      astilectron.sendMessage({ name: "logDisp", payload: "" }, function (message) {
      });
    });
    $("#ctxMenu span.showPollingList").on("click", () => {
      astilectron.sendMessage({ name: "showPollingList", payload: "" }, function (message) {
      });
    });
    $("#ctxMenu span.checkAllPoll").on("click", () => {
      checkAllPoll();
    });
  }
  lastMouseX = mouseX;
  lastMouseY = mouseY;
  return true;
}

function mouseClicked() {
  if (ctxMenu) {
    ctxMenu.remove();
    ctxMenu = undefined;
    return true;
  }
  return false;
}

function mouseReleased() {
  if (draggedNode == "" || !nodes[draggedNode]) {
    draggedNode = "";
    return
  }
  astilectron.sendMessage({ name: "updateNode", payload: nodes[draggedNode] }, function (message) {
  });
  draggedNode = "";
}

function keyReleased() {
  if (!focused) {
    return false;
  }
  if (keyCode == DELETE) {
    // Delete
    deleteNode();
  }
  return true;
}

function deleteNode() {
  if (!selectNode || !nodes[selectNode]) {
    return;
  }
  if (!confirm(`${nodes[selectNode].Name}を削除しますか?`)) {
    return;
  }
  astilectron.sendMessage({ name: "deleteNode", payload: selectNode }, function (message) {
    if (message.payload != "ok") {
      return;
    }
    for (let k in lines) {
      if (lines[k].Node1 == selectNode || lines[k].Node2 == selectNode) {
        delete lines[k];
      }
    }
    delete nodes[selectNode];
    selectNode = "";
    updateNodeList();
  });
}

function dupNode() {
  if (!selectNode) {
    return;
  }
  astilectron.sendMessage({ name: "dupNode", payload: selectNode }, function (message) {
    if (message.payload == "ng") {
      return;
    }
    nodes[message.payload.ID] = message.payload;
    updateNodeList();
  });
}

let log;

function addOrUpdateNode(n) {
  const node = $(`li.list-group-item[data-id=${n.ID}]`);
  const keyword = `${n.State}:${n.Name}:${n.IP}`.replace(`"`, ``);
  if (node.length > 0) {
    $(node).find("i").attr('class', `fas fa-${n.Icon} state state_${n.State}`);
    $(node).find(".media-body strong").html(n.Name);
    $(node).find(".media-body p").html(`${n.IP} ${n.Descr}`);
    $(node).attr('data-keyword',keyword);
  } else {
    const newnode = `
      <li class="list-group-item" data-id="${n.ID}" data-keyword="${keyword}">
        <div class="media-object pull-left">
            <i class="fas fa-${n.Icon} state state_${n.State}"></i>
        </div>
        <div class="media-body">
          <strong>${n.Name}</strong>
          <p>${n.IP} ${n.Descr}</p>
        </div>
      </li>`
    $('#nodeList').append(newnode);
  }
}

function updateNodeList() {
  $('#nodeList li.list-group-item').each((i, e) => {
    const id = $(e).data('id') + '';
    if (!nodes[id]) {
      $(e).remove();
    } else if (id == selectNode) {
      $(e).addClass("selected");
    } else {
      $(e).removeClass("selected");
      $(e).click(id, (e) => {
        const id = e.data;
        if (id != selectNode) {
          selectNode = id;
          updateNodeList();
        }
      });
    }
  });
  clearStatus();
  for (let k in nodes) {
    updateStatus(nodes[k]);
  }
  redraw();
  showStatus();
}


document.addEventListener('astilectron-ready', function () {
  function nodeFilter() {
    const text = $('#nodeFilter').val();
    if ("" == text) {
      $('li[data-keyword]').show();
      return;
    }

    $('li[data-keyword]').hide();
    $('li[data-keyword*=' + text + ']').show();
  }
  $('#nodeFilter').keyup(function () {
    nodeFilter();
    return (false);
  });
  $("header.toolbar-header button.mapConf").on("click", () => {
    createMapConfPane();
  });
  $("header.toolbar-header button.notifyConf").on("click", () => {
    createNotifyConfPane();
  });
  $("header.toolbar-header button.actionPane").on("click", () => {
    createActionPane();
  });
  $("header.toolbar-header button.mibDBConf").on("click", () => {
    createMIBDBPane();
  });
  $("header.toolbar-header button.showPollingList").on("click", () => {
    astilectron.sendMessage({ name: "showPollingList", payload: "" }, function (message) {
    });
  });
  $("header.toolbar-header button.logDisp").on("click", () => {
    astilectron.sendMessage({ name: "logDisp", payload: "" }, function (message) {
    });
  });
  $("header.toolbar-header button.checkAllPoll").on("click", () => {
    checkAllPoll();
  });

  log = $('#log_table').DataTable({
    "order": [[1, "desc"]],
    "paging": false,
    "info": false,
    "autoWidth": true,
    scrollY: 200,
    scrollCollapse: true,
    "language": {
      "search": "フィルター:"
    },
  });  
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "mapConf": {
        mapConf = message.payload;
        setWindowTitle();
        if(mapConf.BackImg ){
          backimg = loadImage("./images/backimg");
        } else {
          backimg = undefined;
        }
        return { name: "mapConf", payload: "ok" };
      }
      case "notifyConf": {
        notifyConf = message.payload;
        return { name: "notifyConf", payload: "ok" };
      }
      case "nodes": {
        nodes = message.payload;
        setTimeout(() => {
          for (let k in nodes) {
            addOrUpdateNode(nodes[k]);
          }
          updateNodeList();
        }, 100);
        return { name: "nodes", payload: "ok" };
      }
      case "lines": {
        lines = message.payload;
        setTimeout(() => {
          redraw();
        }, 100);
        return { name: "nodes", payload: "ok" };
      }
      case "logs": {
        for (let i = message.payload.length - 1; i >= 0; i--) {
          const l = message.payload[i]
          const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
          const lvl = getStateHtml(l.Level)
          log.row.add([lvl, ts, l.Type, l.NodeName, l.Event]);
        }
        if(mapConf && mapConf.LogDispSize){
          // Logの表示数調整
          while( log.rows().count() > mapConf.LogDispSize){
            log.rows(0).remove();
          }
        }
        log.draw();
        return { name: "logs", payload: "ok" };
      }
      case "about": {
        setTimeout(() => {
          astilectron.showMessageBox({ message: message.payload, title: "TWSNMPについて" });
        }, 100);
        return { name: "about", payload: "ok" };
      }
      case "error": {
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
      }
      default:
        console.log(message.name)
        console.log(message.payload)
    }
  });
});

function checkAllPoll() {
  if (!confirm(`全てのノードの再確認を実施しますか?`)) {
    return;
  }
  astilectron.sendMessage({ name: "checkAllPoll", payload: "" }, function (message) {
  });
}

function setWindowTitle() {
  const t = "TWSNMP - " + mapConf.MapName;
  $("title").html(t);
}

function clearStatus() {
  status.High = 0;
  status.Low = 0;
  status.Warn = 0;
  status.Normal = 0;
  status.Unkown = 0;
  status.Repair = 0;
}

function updateStatus(n) {
  switch (n.State) {
    case "high":
      status.High++;
      break;
    case "low":
      status.Low++;
      break;
    case "warn":
      status.Warn++;
      break;
    case "normal":
      status.Normal++;
      break;
    case "repair":
      status.Repair++;
      break;
    default:
      status.Unkown++;
  }
}
function showStatus() {
  const s = "重度=" + status.High + " 軽度=" + status.Low + " 注意=" + status.Warn +
            " 正常=" + status.Normal + " 復帰=" + status.Repair + " 不明="+ status.Unkown;
  $("#status").html(s);
}