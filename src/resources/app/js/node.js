'use strict';

let nodeID = "";
let currentPage = "";
let basic;
let polling;
let template;
let log;
let pane;

let pollingList = {};
let templateList = {};

function setupBasicPage() {
  astilectron.sendMessage({ name: "getNodeBasicInfo", payload: nodeID }, message => {
    if (!message.payload) {
      dialog.showErrorBox("ノード情報", "ノード情報を取得できません。");
      return;
    }
    basic.clear();
    const node = message.payload;
    basic.row.add(["名前", node.Name]);
    basic.row.add(["種別", node.Type]);
    basic.row.add(["IPアドレス", node.IP]);
    basic.row.add(["MACアドレス", node.MAC]);
    basic.row.add(["状態", getStateHtml(node.State)]);
    basic.row.add(["説明", node.Descr]);
    basic.row.add(["Community", node.Community]);
    basic.draw();
    setWindowTitle(node.Name);
  });
}

function setupPollingPage() {
  astilectron.sendMessage({ name: "getNodePollings", payload: nodeID }, message => {
    if ( !message.payload) {
      dialog.showErrorBox("ノード情報", "ポーリングを取得できません。");
      return;
    }
    polling.clear();
    pollingList = {};
    for (let i = message.payload.length - 1; i >= 0; i--) {
      const p = message.payload[i];
      const lt = moment(p.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const level = getStateHtml(p.Level);
      const state = getStateHtml(p.State);
      const logMode = getLogModeHtml(p.LogMode);
      polling.row.add([state, p.Name, level, p.Type,logMode, p.Polling,lt,p.LastVal,p.LastResult, p.ID]);
      pollingList[p.ID] = p;
    }
    polling.draw();
    setPollingBtns(false);
  });
}

function setupLogPage() {
  astilectron.sendMessage({ name: "getNodeLog", payload: nodeID }, message => {
    if (!message.payload[0].Time) {
      dialog.showErrorBox("ノード情報", "ログを取得できません。");
      return;
    }
    log.clear();
    for (let i = message.payload.length - 1; i >= 0; i--) {
      const l = message.payload[i]
      const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const lvl = getStateHtml(l.Level)
      log.row.add([lvl, ts, l.Type, l.Event]);
    }
    log.draw();
  });
}

function showPage(mode) {
  const pages = ["basic", "polling", "log"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $(".toolbar-footer ." + p + "_btns").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $(".toolbar-footer ." + p + "_btns").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  currentPage = mode;
}

function makeBasicTable() {
  basic = $('#basic_table').DataTable({
    "paging": false,
    "info": false,
    "ordering": false,
    "searching": false,
    "autoWidth": true,
  });
}

function setPollingBtns(show,bAn){
  const btns = ["edit","delete","poll"];
  btns.forEach( b =>{
    if(!show) {
      $('.polling_btns button.'+ b).addClass("hidden");
    } else {
      $('.polling_btns button.'+ b).removeClass("hidden");
    }
  });
  if (bAn) {
    $('.polling_btns button.show').removeClass("hidden");
  } else {
    $('.polling_btns button.show').addClass("hidden");
  }
}

function makePollingTable() {
  polling = $('#polling_table').DataTable({
    "paging": true,
    "info": false,
    "order": [[0, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal":        "",
      "emptyTable":     "ポーリングがありません。",
      "info":           "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty":      "",
      "infoFiltered":   "(全 _MAX_ 件)",
      "infoPostFix":    "",
      "thousands":      ",",
      "lengthMenu":     "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing":     "処理中...",
      "search":         "フィルター:",
      "zeroRecords":    "一致するポーリングがありません。",
      "paginate": {
        "first": "最初",
        "last": "最後",
        "next": "次へ",
        "previous": "前へ"
      }, 
      "aria": {
          "sortAscending":  ": 昇順でソート",
          "sortDescending": ": 降順でソート"
      }
    },
  });
  $('#polling_table tbody').on('dblclick', 'tr', function () {
    const data = polling.row( this ).data();
    if(data && data.length > 1){
      const id = data[data.length-1];
      if(pollingList[id] && pollingList[id].LogMode ){
        showPolling(id);
      }
    }
  });
  $('#polling_table tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
      setPollingBtns(false,false);
    } else {
      polling.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
      const r = polling.row('.selected');
      let bAn = false;
      if( r ){
        const d = r.data();
        if (d ){
          const id = d[d.length-1];
          if(pollingList[id] && pollingList[id].LogMode ){
            bAn = true;
          }
        }
      }
      setPollingBtns(true,bAn);
    }
  });
  $('.polling_btns button.delete').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    if (!confirmDialog("ノード削除",`${pollingList[id].Name}を削除しますか?`)) {
      return;
    }
    astilectron.sendMessage({ name: "deletePolling", payload: id }, message => {
      if (message.payload != "ok" ) {
        dialog.showErrorBox("ポーリング削除", "削除できません。");
        return;
      }
      const r = polling.row('.selected');
      if (r) {
        r.remove().draw(false);
      }
    });
  });
  $('.polling_btns button.poll').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    astilectron.sendMessage({ name: "pollNow", payload: id }, message => {
      if (message.payload != "ok" ) {
        dialog.showErrorBox("ポーリング確認", "ポーリングの再実行に失敗しました。");
        return;
      }
      setupPollingPage();
      showPage("polling");
    });
  });
  $('.polling_btns button.edit').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    createEditPollingPane(id,"");
  });
  $('.polling_btns button.add').click(function () {
    showTemplate();
  });
  $('.polling_btns button.auto').click(function () {
    autoAddPolling();
  });
  $('.polling_btns button.show').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    showPolling(id);
  });
}

function showPolling(id) {
  astilectron.sendMessage({ name: "showPolling", payload: id }, message => {
    if (message.payload != "ok" ) {
      dialog.showErrorBox("ノード情報", "ポーリング分析画面を表示できません。");
      return;
    }
  });
}

function getSelectedPollingID() {
  const r = polling.row('.selected');
  if (!r) {
    return undefined;
  }
  const d = r.data();
  if (!d) {
    return undefined;
  }
  const id = d[d.length-1];
  if (!pollingList[id]) {
    return undefined;
  }
  return id
}


function makeLogTable() {
  log = $('#log_table').DataTable({
    "paging": true,
    "info": false,
    "order": [[1, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal":        "",
      "emptyTable":     "表示するログがありません。",
      "info":           "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty":      "",
      "infoFiltered":   "(全 _MAX_ 件)",
      "infoPostFix":    "",
      "thousands":      ",",
      "lengthMenu":     "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing":     "処理中...",
      "search":         "フィルター:",
      "zeroRecords":    "一致するログがありません。",
      "paginate": {
          "first":      "最初",
          "last":       "最後",
          "next":       "次へ",
          "previous":   "前へ"
      },
      "aria": {
          "sortAscending":  ": 昇順でソート",
          "sortDescending": ": 降順でソート"
      }
    },
  });
}

document.addEventListener('astilectron-ready', function () {
  makeBasicTable();
  makePollingTable();
  makeLogTable();
  makeTemplateTable();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "setNodeID":
        if (message.payload) {
          nodeID = message.payload;
          setupBasicPage();
          showPage("basic");
        }
        return { name: "setNodeID", payload: "ok" };
      case "setMode":
        if (message.payload) {
          if(message.payload == "showNodeLog"){
            setupLogPage()
            showPage("log")
          } else if (message.payload == "showPolling"){
            setupPollingPage()
            showPage("polling")
          }
        }
        return { name: "setNodeID", payload: "ok" };
      case "error":
        setTimeout(() => {
          dialog.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
  $('#basic').click(() => {
    if(pane){
      return true;
    }
    setupBasicPage();
    showPage("basic");
  });
  $('#polling').click(() => {
    if(pane){
      return true;
    }
    setupPollingPage();
    showPage("polling");
  });
  $('#log').click(() => {
    if(pane){
      return true;
    }
    setupLogPage();
    showPage("log");
  });
});

function createEditPollingPane(id,tid){
  if(pane){
    pane.dispose();
    pane = undefined;
  }
  let p;
  if( id && pollingList[id] ){
    p = pollingList[id];
  } else {
    p = {
      ID: "",
      Name: "",
      NodeID: nodeID,
      Type: "ping",
      Polling: "",
      Level: "low",
      PollInt:   60,
      Timeout: 1,
      Retry: 1,
      NextTime:0,
      LogMode: 0,
      LastTime: 0,
      LastResult: "",
      State: "unknown",
    };
    if(tid) {
      p.Name = templateList[tid].Name;
      p.Type = templateList[tid].Type;
      p.Level = templateList[tid].Level;
      p.Polling = templateList[tid].Polling;
    }
  }
  pane = new Tweakpane({
    title: id === "" ? "新規ポーリング" : "ポーリング編集",
  });
  pane.addInput(p, 'Name', { label: "名前" });
  pane.addInput(p, 'Type', { 
    label: "種別",
    options: pollingTypeList,
  });
  pane.addInput(p, 'Level', { 
    label: "レベル",
    options: levelList,
  });
  pane.addInput(p, 'Polling', { label: "定義" });
  pane.addInput(p, 'PollInt', { 
    label: "間隔",
    min: 60,
    max: 600,
    step: 10,
  });
  pane.addInput(p, 'Timeout', { 
    label: "Timeout",
    min: 1,
    max: 5,
    step: 1,
  });
  pane.addInput(p, 'Retry', { 
    label: "Retry",
    min: 0,
    max: 5,
    step: 1,
  });
  pane.addInput(p, 'LogMode', { 
    label: "ログモード",
    options: logModeList,
  });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
    pane = undefined;
  });
  pane.addButton({
    title: 'Save',
  }).on('click', (value) => {
    clearInputError();
    if( p.Name == "" ){
      setInputError(0,"名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "savePolling", payload: p }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("ポーリング編集", "保存に失敗しました。");
        return;
      }
      setupPollingPage();
      showPage("polling");
    });
    pane.dispose();
    pane = undefined;
  });
}

function setWindowTitle(n){
  const t = "ノード情報 - " + n;
  $("title").html(t);
  $("h1.title").html(t);
}


function makeTemplateTable() {
  template = $('#template_table').DataTable({
    "paging": true,
    "info": false,
    "order": [[0, "desc"]],
    "searching": true,
    "autoWidth": true,
    "language": {
      "decimal":        "",
      "emptyTable":     "テンプレートがありません。",
      "info":           "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty":      "",
      "infoFiltered":   "(全 _MAX_ 件)",
      "infoPostFix":    "",
      "thousands":      ",",
      "lengthMenu":     "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing":     "処理中...",
      "search":         "フィルター:",
      "zeroRecords":    "一致するテンプレートがありません。",
      "paginate": {
        "first": "最初",
        "last": "最後",
        "next": "次へ",
        "previous": "前へ"
      }, 
      "aria": {
          "sortAscending":  ": 昇順でソート",
          "sortDescending": ": 降順でソート"
      }
    },
  });
  $('#template_table tbody').on('dblclick', 'tr', function () {
    const data = polling.row( this ).data();
    if(data && data.length > 1){
      const id = data[data.length-1];
      if(templateList[id] ){
        $('select_template').click();
      }
    }
  });
  $('#template_table tbody').on('click', 'tr', function () {
    if ($(this).hasClass('selected')) {
      $(this).removeClass('selected');
    } else {
      template.$('tr.selected').removeClass('selected');
      $(this).addClass('selected');
    }
  });
}

function getSelectedTemplateID() {
  const r = template.row('.selected');
  if (!r) {
    return undefined;
  }
  const d = r.data();
  if (!d) {
    return undefined;
  }
  const id = d[d.length-1];
  if (!templateList[id]) {
    return undefined;
  }
  return id
}

function showTemplate() {
  astilectron.sendMessage({ name: "getTemplates", payload: "" }, message => {
    if (!message.payload || message.payload == "ng" ) {
      createEditPollingPane("","");
      return;
    }
    template.clear();
    templateList =  message.payload;
    for(let id in templateList){
      const t = templateList[id];
      const level = getStateHtml(t.Level);
      template.row.add([level,t.Type,t.Name,t.Polling,t.NodeType,t.Descr,t.ID]);
    }
    template.draw();
    $("#template_win").addClass("show").fadeIn().css('display', 'flex');
    $("#cancel_template").on("click", function () {
      $("#template_win").fadeOut();
    });
    $("#select_template").on("click", function () {
      $("#template_win").fadeOut();
      const tid = getSelectedTemplateID();
      createEditPollingPane("",tid);
    });
  });
}

function autoAddPolling() {
  astilectron.sendMessage({ name: "autoAddPolling", payload: nodeID }, message => {
    if (!message.payload || message.payload == "ng") {
      return;
    }
    setupPollingPage();
  });
}
