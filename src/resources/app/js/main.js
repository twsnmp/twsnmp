'use strict';

let myFont;
let selectNode = "";
let mapConf;
let nodes = {};
let lines = {};

// ICONS
const iconArray =[
  ["desktop",0xf108],
  ["tablet",0xf3fa],
  ["server",0xf233],
  ["hdd",0xf0a0],
  ["laptop",0xf109],
  ["network-wired",0xf6ff],
  ["wifi",0xf1eb],
  ["cloud",0xf0c2],
  ["print",0xf02f]
];
const iconMap = new Map(iconArray);

// State Colors
const stateColorArray = [
    ["high","#e31a1c"],
    ["low","#fb9a99"],
    ["warn","#dfdf22"],
    ["normal","#33a02c"],
    ["info","#1f78b4"],
    ["repair","#1f78b4"]
];
const  stateColorMap = new Map(stateColorArray);

// State Html
const stateHtmlArray = [
  ["high",'<i class="fas fa-exclamation-circle state state_high"></i>重度'],
  ["low",'<i class="fas fa-exclamation-circle state state_low"></i>軽度'],
  ["warn",'<i class="fas fa-exclamation-triangle state state_warn"></i>注意'],
  ["normal",'<i class="fas fa-check-circle state state_normal"></i>正常'],
  ["info",'<i class="fas fa-info-circle state state_info"></i>情報'],
  ["repair",'<i class="fas fa-check-circle state state_repair"></i>復帰']
];

const  stateHtmlMap = new Map(stateHtmlArray);

function getIcon(icon) {
  const ret = iconMap.get(icon);
  return  ret  ? char(ret) : char(0xf059);
 }

function getStateColor(state) {
  const ret = stateColorMap.get(state);
  return  ret ? color(ret) : color("#999");
}

function getStateHtml(state) {
  const ret = stateHtmlMap.get(state);
  return  ret ? ret : '<i class="fas fa-check-circle state state_unkown"></i>不明';
}

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
  for(let k in lines) {
    if (!nodes[lines[k].NodeID1] || !nodes[lines[k].NodeID2] ) {
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
  for(let k in nodes){
    const icon = getIcon(nodes[k].Icon);
    push();
    translate(nodes[k].X, nodes[k].Y);
    if (selectNode == nodes[k].ID) {
      fill("aliceblue");
      stroke(getStateColor(nodes[k].State));
      rect(-24, -24, 48, 48);
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

function seletcNode() {
  for(let k in nodes) {
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
    winMouseY < 55 ||
    winMouseY > windowHeight * 0.75) {
    return true;
  }
  if ( nodes[selectNode] && lastMouseX) {
    nodes[selectNode].X += mouseX - lastMouseX;
    nodes[selectNode].Y += mouseY - lastMouseY;
    if( nodes[selectNode].X < 16){
      nodes[selectNode].X = 16;
    }
    if( nodes[selectNode].Y < 16){
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
    winMouseY < 55 ||
    winMouseY > windowHeight * 0.75) {
    return true;
  }
  if(ctxMenu){
    return true;
  }

  const selectNodeBack = selectNode;
  seletcNode();
  if(keyIsDown(SHIFT) && 
    selectNodeBack != "" &&
    selectNode != "" && 
    selectNodeBack != seletcNode) {
    astilectron.sendMessage({ name: "editLine", payload: {NodeID1:selectNodeBack,NodeID2:selectNode} }, function (message) {
    });
    selectNode = "";
    return true;
  }
  if (selectNodeBack != seletcNode) {
    updateNodeList();
    redraw();
  }
  if (mouseButton === RIGHT) {
    let div;
    if (nodes[selectNode]) {
      div =`
      <nav class="nav-group">
        <span class="nav-group-item showNodeInfo">
          <i class="fas fa-info-circle"></i>    
         情報
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
          削除
        </span>
      </nav>
      `;
    } else {
      div =`
      <nav class="nav-group">
        <span class="nav-group-item startDiscover">
          <i class="fas fa-search"></i>
          自動発見
        </span>
        <span class="nav-group-item addNode">
          <i class="fas fa-plus-circle"></i>
          新規ノード
        </span>
        <span class="nav-group-item configMap">
          <i class="fas fa-cog"></i>
          マップ設定
        </span>
      </nav>
      `;    
    }
    ctxMenu = createDiv(div);
    ctxMenu.id("ctxMenu");
    ctxMenu.position(winMouseX,winMouseY + 10);
    $("#ctxMenu span.deleteNode").on("click",()=>{
      deleteNode();
    });  
    $("#ctxMenu span.dupNode").on("click",()=>{
      dupNode();
    });  
    $("#ctxMenu span.showNodeInfo").on("click",()=>{
      if(selectNode != "") {
        astilectron.sendMessage({ name: "editNode", payload: selectNode }, function (message) {
        });
      }
    });  
    $("#ctxMenu span.editNode").on("click",()=>{
      if(selectNode != "") {
        astilectron.sendMessage({ name: "editNode", payload: selectNode }, function (message) {
        });
      }
    });  
    $("#ctxMenu span.startDiscover").on("click",()=>{
      astilectron.sendMessage({ name: "startDiscover", payload: {X:lastMouseX,Y:lastMouseY} }, function (message) {
      });
    });  
    $("#ctxMenu span.addNode").on("click",()=>{
      astilectron.sendMessage({ name: "addNode", payload: {X:lastMouseX,Y:lastMouseY} }, function (message) {
      });
    });  
    $("#ctxMenu span.configMap").on("click",()=>{
      astilectron.sendMessage({ name: "configMap", payload: "" }, function (message) {
      });
    });  
  }
  lastMouseX = mouseX;
  lastMouseY = mouseY;
  return true;
}

function mouseClicked(){
  if(ctxMenu){
    ctxMenu.remove();
    ctxMenu = undefined;
    return true;
  }
  return false;
}

function mouseReleased() {
  if( draggedNode == "" || !nodes[draggedNode]) {
    draggedNode = "";
    return
  }
  astilectron.sendMessage({ name: "updateNode", payload: nodes[draggedNode] }, function (message) {
  });
  draggedNode = "";
}

function keyReleased() {
  if(!focused){
    return false;
  }
  if(keyCode == DELETE ) {
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
    if( message.payload != "ok") {
      return;
    }
    for(let k in lines){
      if(lines[k].Node1 == selectNode || lines[k].Node2 == selectNode){
        delete lines[k];
      }
    }
    delete nodes[selectNode];
    selectNode = "";
    updateNodeList();
    redraw();
  });
}

function dupNode() {
  if (!selectNode) {
    return;
  }
  astilectron.sendMessage({ name: "dupNode", payload: selectNode }, function (message) {
    if( message.payload == "ng") {
      return;
    }
    nodes[message.payload.ID] = message.payload;
    updateNodeList();
    redraw();
  });
}

let log;

function addOrUpdateNode(n) {
  const node = $(`li.list-group-item[data-id=${n.ID}]`);
  if (node.length > 0) {
    $(node).find("i").attr('class', `fas fa-${n.Icon} state state_${n.State}`);
    $(node).find("media-body strong").html(n.Name);
    $(node).find("media-body p").html(n.Descr);
  } else {
    const keyword = `${n.State}:${n.Name}`.replace(`"`, ``);
    const newnode = `
      <li class="list-group-item" data-id="${n.ID}" data-keyword="${keyword}">
        <div class="media-object pull-left">
            <i class="fas fa-${n.Icon} state state_${n.State}"></i>
        </div>
        <div class="media-body">
          <strong>${n.Name}</strong>
          <p>${n.Descr}</p>
        </div>
      </li>`
    $('#nodeList').append(newnode);
  }
}

function updateNodeList() {
  $('#nodeList li.list-group-item').each((i, e) => {
    const id = $(e).data('id') +'';
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
          redraw();
        }
      });
    }
  });
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
    return(false);
  });
  log = $('#log_table').DataTable({
    "order": [[1,"desc"]],
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
        return { name: "mapConf", payload: "ok" };
      }
      case "nodes": {
        nodes = message.payload;
        setTimeout(() => {
          for(let k in nodes) {
            addOrUpdateNode(nodes[k]);
          }
          updateNodeList();
          redraw();
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
      case "logs":{
        for(let i= message.payload.length-1 ; i >= 0 ;i--){
          const l = message.payload[i]
          const ts = moment(l.Time/(1000*1000)).format("MM/DD HH:mm:ss.SSS");
          const lvl = getStateHtml(l.Level)
          log.row.add([lvl,ts,l.Type,l.NodeName,l.Event]);
        }
        log.draw();
        return { name: "logs", payload: "ok" };
      }
      case "about":{
        setTimeout(() => {
          astilectron.showMessageBox({ message: message.payload, title: "TWSNMPについて" });
        }, 100);
        return { name: "about", payload: "ok" };
      }
      case "error":{
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