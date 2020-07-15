'use strict';

let myFont;
let nodes;
let lines;
let pollings;
let backimg;
let username;
let password;

let log; // Table
let polling; // Tabale

$(document).ready(function () {
  $("#login_form").on('keyup',function (e) {
    if (e.keyCode == 13) {
      $("#login_btn").click();
    }
  });
  $('#login_btn').on('click', () => {
    $.ajax({
      url: '/api/mapstatus',
      type: 'GET',
      username: $('#signin_form [name=username]').val(),
      password: $('#signin_form [name=password]').val()
    })
    .done((r) => {
      loginOK(r)
    })
    // Ajaxリクエストが失敗した時発動
    .fail((r) => {
      $('#login_err').html("ユーザー名またはパスワードが違います。");
      console.log(r)
    });
  });
  $('#nodeFilter').keyup(function () {
    nodeFilter();
    return (false);
  });
  makeLogTable();
  makePollingTable();
});

function loginOK(r) {
  username = $('#login_form [name=username]').val();
  password = $('#login_form [name=password]').val();
  $('div.tab-group').removeClass("hidden");
  $('#btn-group').removeClass("hidden");
  $('#login_page').addClass("hidden");
  $('#map_tab').on('click', () => {
    showMapPage();
  });
  $('#polling_tab').on('click', () => {
    showPollingPage();
  });
  $('#logout_btn').on('click', () => {
    logOut();
  });
  $('#reload_btn').on('click', () => {
    updateMapData();
  });
  showMapPage();
  showStatus(r);
  updateMapData();
}

function updateMapData() {
  if(!username){
    return;
  }
  $.ajax({
    url: '/api/mapdata',
    type: 'GET',
    headers: {
      "Authorization": "Basic " + btoa(username + ":" + password)
    },
    username: username,
    password: password
  })
  .done((r) => {
    setMapData(r)
  })
  // Ajaxリクエストが失敗した時発動
  .fail((r) => {
    console.log(r)
  });
  $.ajax({
    url: '/api/mapstatus',
    type: 'GET',
    headers: {
      "Authorization": "Basic " + btoa(username + ":" + password)
    },
    username: username,
    password: password
  })
  .done((r) => {
    showStatus(r);
  })
  // Ajaxリクエストが失敗した時発動
  .fail((r) => {
    console.log(r)
  });
}

function logOut() {
  username = "";
  password = "";
  $('#login_form [name=username]').val("");
  $('#login_form [name=password]').val("");

  $('div.tab-group').addClass("hidden");
  $('#btn-group').addClass("hidden");
  $('#map_page').addClass("hidden");
  $('#polling_page').addClass("hidden");
  $('#login_page').removeClass("hidden");
  $('#title').html("TWSNMPログイン");
  $('title').html("TWSNMPログイン");
  $('#status').html('');
}

function showMapPage() {
  $('#map_page').removeClass("hidden");
  $('#polling_page').addClass("hidden");
  $("#map_tab").addClass("active");
  $("#polling_tab").removeClass("active");
}

function setMapData(r) {
  nodes   = r.Nodes;
  lines   = r.Lines;
  setWindowTitle(r.MapName);
  if(r.BackImg){
    loadImage("/images/backimg",img => {
      backimg =  img;
    });
  } else {
    backimg = undefined;
  }
  updateLogTable(r.Logs);
  updatePollingTable(r.Pollings);
  $('#nodeList').html('');
  for (let k in nodes) {
    addNodeToList(nodes[k]);
  }
  redraw();
}

function showPollingPage() {
  $('#map_page').addClass("hidden");
  $('#polling_page').removeClass("hidden");
  $("#map_tab").removeClass("active");
  $("#polling_tab").addClass("active");
}


function addNodeToList(n) {
  const node = $(`li.list-group-item[data-id=${n.ID}]`);
  const keyword = `${n.State}:${n.Name}:${n.IP}`.replace(`"`, ``);
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


function updateLogTable(logs) {
  log.clear();
  for (let i = logs.length - 1; i >= 0; i--) {
    const l = logs[i]
    const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    const lvl = getStateHtml(l.Level)
    log.row.add([lvl, ts, l.Type, l.NodeName, l.Event]);
  }
  log.draw();
}

function updatePollingTable(pollings) {
  polling.clear();
  for (let i = pollings.length - 1; i >= 0; i--) {
    const p = pollings[i];
    let nodeName = "unkown";
    if (p.NodeID && nodes[p.NodeID] ) {
      nodeName = nodes[p.NodeID].Name;
    }
    const lt = moment(p.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
    const level = getStateHtml(p.Level);
    const state = getStateHtml(p.State);
    const logMode = getLogModeHtml(p.LogMode);
    polling.row.add([state,nodeName, p.Name, level, p.Type,logMode, p.Polling,lt,p.LastVal,p.LastResult, p.ID]);
  }
  polling.draw();
}

function setWindowTitle(name) {
  const t = "TWSNMP - " + name;
  $("title").html(t);
  $("#title").html(t);
}

function showStatus(mapStatus) {
  let s = "重度=" + mapStatus.High + " 軽度=" + mapStatus.Low + " 注意=" + mapStatus.Warn +
  " 正常=" + mapStatus.Normal + " 復帰=" + mapStatus.Repair + " 不明="+ mapStatus.Unkown;
  if( mapStatus.DBStatsStr ){
    s += " DBサイズ=" + mapStatus.DBSizeStr;
  }
  $("#status").html(s);
}

function makeLogTable() {
  log = $('#log_table').DataTable({
    dom: 'Bfrt',
    buttons: [
      {
        extend:    'copyHtml5',
        text:      '<i class="fas fa-copy"></i>',
        titleAttr: 'Copy'
      },
      {
          extend:    'excelHtml5',
          text:      '<i class="fas fa-file-excel"></i>',
          titleAttr: 'Excel'
      },
      {
          extend:    'csvHtml5',
          text:      '<i class="fas fa-file-csv"></i>',
          titleAttr: 'CSV'
      }
    ],
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
}

function makePollingTable() {
  polling = $('#polling_table').DataTable({
    dom: 'lBfrtip',
    buttons: [
      {
        extend:    'copyHtml5',
        text:      '<i class="fas fa-copy"></i>',
        titleAttr: 'Copy'
      },
      {
          extend:    'excelHtml5',
          text:      '<i class="fas fa-file-excel"></i>',
          titleAttr: 'Excel'
      },
      {
          extend:    'csvHtml5',
          text:      '<i class="fas fa-file-csv"></i>',
          titleAttr: 'CSV'
      }
    ],
    "paging": true,
    "pageLength": 25,
    "info": false,
    "order": [[0, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal": "",
      "emptyTable": "表示するポーリングがありません。",
      "info": "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty": "",
      "infoFiltered": "(全 _MAX_ 件)",
      "infoPostFix": "",
      "thousands": ",",
      "lengthMenu": "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing": "処理中...",
      "search": "フィルター:",
      "zeroRecords": "一致するポーリングがありません。",
      "paginate": {
        "first": "最初",
        "last": "最後",
        "next": "次へ",
        "previous": "前へ"
      },
      "aria": {
        "sortAscending": ": 昇順でソート",
        "sortDescending": ": 降順でソート"
      }
    },
  });
}

function nodeFilter() {
  const text = $('#nodeFilter').val();
  if ("" == text) {
    $('li[data-keyword]').show();
    return;
  }
  $('li[data-keyword]').hide();
  $('li[data-keyword*="' + text + '"]').show();
}


// P5.JS Functions
function preload() {
  myFont = loadFont('/webfonts/fa-solid-900.ttf');
}

function setup() {
  var canvas = createCanvas(2500, 5000);
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
    fill('rgba(250,250,250,0.8)')
    stroke(250);
    rect(-18, -18, 36, 36);
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
  ["print",0xf02f],
  ["sync",0xf021],
  ["mobile-alt",0xf3cd],
  ["tv",0xf26c],
  ["database",0xf1c0],
  ["clock",0xf017],
  ["phone",0xf095],
  ["video",0xf03d],
  ["globe",0xf0ac],
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

const logModeHtml = [
  '<i class="fas fa-stop-circle state state_unknown"></i>しない',
  '<i class="fas fa-video state state_info"></i>常時',
  '<i class="fas fa-ellipsis-h state state_info"></i>変化時',
  '<i class="fas fa-brain state state_high"></i>AI分析',
];

function getLogModeHtml(m) {
  if( m >=0 && m < logModeHtml.length){
    return logModeHtml[m];
  }
  return  logModeHtml[0];
}
