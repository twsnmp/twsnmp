'use strict';

let mode = "";
let icon = "desktop";

function showPage(p) {
  const pageList = ["configMap", "editNode", "editLine","startDiscover","discoverStat"];
  mode = "";
  pageList.forEach(e => {
    if (p === e) {
      mode = p;
      $("#" + e + "_page").removeClass("hidden");
    } else {
      $("#" + e + "_page").addClass("hidden");
    }
  });
}

function setupConfigMapDlg(c) {
  $("title").html("マップ設定");
  $("h1.title").html("マップ設定");
  $("#configMap_form [name=mapname]").val(c.MapName);
  $("#configMap_form [name=pollint]").val(c.PollInt);
  $("#configMap_form [name=timeout]").val(c.Timeout);
  $("#configMap_form [name=retry]").val(c.Retry);
  $("#configMap_form [name=logdispsize]").val(c.LogDispSize);
  $("#configMap_form [name=logdays]").val(c.LogDays);
  $("#configMap_form [name=community]").val(c.Community);
  $("#configMap_form [name=syslog]").prop("checked", c.EnableSyslogd);
  $("#configMap_form [name=trap]").prop("checked", c.EnableTrapd);
  $("#configMap_form [name=netflow]").prop("checked", c.EnableNetflowd);
  $("#configMap_form [name=backimg]").val(c.BackImg);
  showPage("configMap");
}

function setupEditNodeDlg(c) {
  if(c.Icon){
    icon = c.Icon;
  }
  $("title").html("ノード設定");
  $("h1.title").html("ノード設定");
  $("#editNode_form [name=name]").val(c.Name);
  $('#editNode_form [data-icon="'+icon+'"]').addClass("active");
  $("#editNode_form [name=descr]").val(c.Descr);
  $("#editNode_form [name=ip]").val(c.IP);
  $("#editNode_form [name=community]").val(c.Community);
  $("#editNode_form [name=x]").val(c.X);
  $("#editNode_form [name=y]").val(c.Y);
  $("#editNode_form [name=id]").val(c.ID);
  showPage("editNode");
  $("#editNode_form .btn-group button.btn").on("click",e=>{
    $("#editNode_form .btn-group button.btn").removeClass("active");
    $(e.currentTarget).addClass("active");
    // $(this).addClass("active");  //thisは、undefinedになる
  });
}

let lastLineData;
function setupEditLineDlg(c) {
  lastLineData = c.Line;
  $("title").html("ライン設定");
  $("h1.title").html("ライン設定");
  $("#editLine_form [name=node1]").val(c.NodeName1);
  $("#editLine_form [name=node2]").val(c.NodeName2);
  $("#editLine_form [name=polling2]").empty();
  $("#editLine_form [name=polling1]").empty();
  c.Pollings1.forEach((p,i)=>{
    let option = $('<option>')
    .val(p.ID)
    .text(p.Name)
    .prop('selected', c.Line.PollingID1 == p.ID );
    $("#editLine_form [name=polling1]").append(option);
  });
  c.Pollings2.forEach((p,i)=>{
    let option = $('<option>')
    .val(p.ID)
    .text(p.Name)
    .prop('selected', c.Line.PollingID2 == p.ID );
    $("#editLine_form [name=polling2]").append(option);
  });
  if(c.Line.ID === ""){
    $('#editLine_form .del_btn').addClass("hidden");
  } else {
    $('#editLine_form .del_btn').removeClass("hidden");
  }
  showPage("editLine");
}

function getConfigMapPayload(){
  const ret = {
    MapName: $("#configMap_form [name=mapname]").val(),
    PollInt: $("#configMap_form [name=pollint]").val() * 1,
    Timeout: $("#configMap_form [name=timeout]").val() * 1,
    Retry: $("#configMap_form [name=retry]").val() * 1,
    LogDispSize: $("#configMap_form [name=logdispsize]").val() * 1,
    LogDays: $("#configMap_form [name=logdays]").val() * 1,
    Community: $("#configMap_form [name=community]").val(),
    EnableSyslogd: $("#configMap_form [name=syslog]").prop("checked"),
    EnableTrapd: $("#configMap_form [name=trap]").prop("checked"),
    EnableNetflowd: $("#configMap_form [name=netflow]").prop("checked"),
    BackImg: $("#configMap_form [name=backimg]").val()
  }
  if (ret.MapName === "") {
    astilectron.showErrorBox("マップ名エラー", "マップ名を指定してください。")
    return;
  }
  return ret
}

function getEditNodePayload(){
  const ret = {
    ID: $("#editNode_form [name=id]").val(),
    Name: $("#editNode_form [name=name]").val(),
    Icon: $("#editNode_form button.active").data("icon"),
    Descr: $("#editNode_form [name=descr]").val(),
    X: $("#editNode_form [name=x]").val()*1,
    Y: $("#editNode_form [name=y]").val()*1,
    IP: $("#editNode_form [name=ip]").val(),
    Community: $("#editNode_form [name=community]").val(),
  }
  if (ret.Name === "") {
    astilectron.showErrorBox("ノード名エラー", "ノード名を指定してください。")
    return;
  }
  if( !ret.Icon){
    ret.Icon = 'desktop';
  }
  return ret;
}

function getEditNodePayload(){
  const ret = {
    ID: $("#editNode_form [name=id]").val(),
    Name: $("#editNode_form [name=name]").val(),
    Icon: $("#editNode_form button.active").data("icon"),
    Descr: $("#editNode_form [name=descr]").val(),
    X: $("#editNode_form [name=x]").val()*1,
    Y: $("#editNode_form [name=y]").val()*1,
    IP: $("#editNode_form [name=ip]").val(),
    Community: $("#editNode_form [name=community]").val(),
  }
  if (ret.Name === "" ) {
    astilectron.showErrorBox("ノード名エラー", "ノード名を指定してください。")
    return;
  }
  if( !ret.Icon){
    ret.Icon = 'desktop';
  }
  return ret;
}

function getEditLinePayload(){
  lastLineData.PollingID1 = $("#editLine_form [name=polling1]").val()
  lastLineData.PollingID2 = $("#editLine_form [name=polling2]").val()
  return lastLineData;
}

let lastStartDiscover;
function setupStartDiscoverDlg(c) {
  lastStartDiscover = c;
  $("title").html("自動発見設定");
  $("h1.title").html("自動発見設定");
  $("#startDiscover_form [name=startip]").val(c.StartIP);
  $("#startDiscover_form [name=endip]").val(c.EndIP);
  $("#startDiscover_form [name=timeout]").val(c.Timeout);
  $("#startDiscover_form [name=retry]").val(c.Retry);
  $("#startDiscover_form [name=community]").val(c.Community);
  showPage("startDiscover");
}

function setupDiscoverStatDlg(c) {
  $("title").html("自動発見状況");
  $("h1.title").html("自動発見状況");
  $("#").val(c.Proggress);
  showPage("discoverStat");
}

function getStartDiscoverPayload(){
  lastStartDiscover.StartIP = $("#startDiscover_form [name=startip]").val();
  lastStartDiscover.EndIP = $("#startDiscover_form [name=endip]").val();
  lastStartDiscover.Community = $("#startDiscover_form [name=community]").val();
  lastStartDiscover.Timeout = $("#startDiscover_form [name=timeout]").val()*1;
  lastStartDiscover.Retry = $("#startDiscover_form [name=retry]").val()*1;
  if (lastStartDiscover.StartIP === "" || lastStartDiscover.EndIP === ""  ) {
    astilectron.showErrorBox("範囲指定エラー", "開始、終了IPアドレスが正しくありません。")
    return;
  }
  return lastStartDiscover;
}

function getPayload() {
  switch (mode) {
    case "configMap":{
      return getConfigMapPayload();
    }
    case "editNode": {
      return getEditNodePayload();
    }
    case "editLine":{
      return getEditLinePayload();
    }
  }
}

document.addEventListener('astilectron-ready', function () {
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "configMap":
        setupConfigMapDlg(message.payload);
        return { name: "configMap", payload: "ok" };
      case "editNode":
        setupEditNodeDlg(message.payload);
        return { name: "editNode", payload: "ok" };
      case "editLine":
        setupEditLineDlg(message.payload);
        return { name: "editLine", payload: "ok" };
      case "startDiscover":
        setupStartDiscoverDlg(message.payload);
        return { name: "startDiscover", payload: "ok" };
      case "discoverStat":
        setupDiscoverStatDlg(message.payload);
        return { name: "discoverStat", payload: "ok" };
      case "error":
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('.save_btn').click(() => {
    const payload = getPayload();
    if (payload) {
      astilectron.sendMessage({ name: "save." + mode, payload: payload }, message => {
      });
    }
  });
  $('.del_btn').click(() => {
    const payload = getPayload();
    if (payload) {
      astilectron.sendMessage({ name: "del." + mode, payload: payload }, message => {
      });
    }
  });
  $('.cancel_btn').click(() => {
    astilectron.sendMessage({ name: "cancel", payload: "" }, message => {
    });
  });
  $('.startDiscover_btn').click(() => {
    astilectron.sendMessage({ name: "startDiscover", payload: getStartDiscoverPayload() }, message => {
    });
  });
  $('.stopDiscover_btn').click(() => {
    astilectron.sendMessage({ name: "stopDiscover", payload: "" }, message => {
    });
  });
  $('#select_backimg').on("click", function () {
    astilectron.showOpenDialog({ properties: ['openFile'], title: "背景画像ファイル" }, function (paths) {
      $("#backimg").val(paths[0]);
    });
  });
  $("#backimg").on("drop", function (e) {
    e.preventDefault();
    if (e.originalEvent.dataTransfer.files.length == 1) {
      $("#backimg").val(e.originalEvent.dataTransfer.files[0].path);
    }
  });
  $("#backimg").on("dragover", function (e) {
    e.preventDefault();
  });
});
