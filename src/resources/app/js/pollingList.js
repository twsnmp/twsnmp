'use strict';

let polling;
let pane;

let pollingList = {};
let nodes = {};

function setupPolling() {
  astilectron.sendMessage({ name: "getPollings", payload: "" }, message => {
    if (!message.payload.Pollings) {
      dialog.showErrorBox("ポーリングリスト", "ポーリングを取得できません。");
      return;
    }
    nodes = message.payload.Nodes
    polling.clear();
    pollingList = {};
    for (let i = message.payload.Pollings.length - 1; i >= 0; i--) {
      const p = message.payload.Pollings[i];
      let nodeName = "unkown";
      if (p.NodeID && nodes[p.NodeID] ) {
        nodeName = nodes[p.NodeID].Name;
      }
      const lt = moment(p.LastTime / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
      const level = getStateHtml(p.Level);
      const state = getStateHtml(p.State);
      const logMode = getLogModeHtml(p.LogMode);
      polling.row.add([state,nodeName, p.Name, level, p.Type,logMode, p.Polling,lt,p.LastVal,p.LastResult, p.ID]);
      pollingList[p.ID] = p;
    }
    polling.draw();
    setPollingBtns(false,false);
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
      "search": "検索:",
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
        if (d){
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
    if (!confirm(`${pollingList[id].Name}を削除しますか?`)) {
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
      setupPolling();
    });
  });
  $('.polling_btns button.edit').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    createEditPollingPane(id);
  });
  $('.polling_btns button.add').click(function () {
    createEditPollingPane("");
  });
  $('.polling_btns button.show').click(function () {
    const id = getSelectedPollingID()
    if(!id){
      return;
    }
    astilectron.sendMessage({ name: "showPolling", payload: id }, message => {
      if (message.payload != "ok" ) {
        dialog.showErrorBox("ポーリングリスト", "ポーリング分析画面を表示できません。");
        return;
      }
    });
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

document.addEventListener('astilectron-ready', function () {
  makePollingTable();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "show":
        setupPolling()
        return { name: "show", payload: "ok" };
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
});

function createEditPollingPane(id){
  if(pane){
    pane.dispose();
    pane = undefined;
  }
  let p;
  if( pollingList[id] ){
    p = pollingList[id];
  } else {
    p = {
      ID: "",
      Name: "",
      NodeID: "",
      Type: "ping",
      Polling: "",
      Level: "low",
      PollInt:   60,
      Timeout: 1,
      Retry: 1,
      NextTime:0,
      DispMode:"",
      LogMode: 0,
      LastTime: 0,
      LastResult: "",
      State: "unkown",
    };
  }
  pane = new Tweakpane({
    title: id === "" ? "新規ポーリング" : "ポーリング編集",
  });
  if(id == "") {
    const opts = {};
    for(let k in nodes){
      const e = nodes[k];
      if(e){
        opts[e.Name] = e.NodeID;
      }
    }
    pane.addInput(p, 'NodeID', { 
      label: "ノード",
      options: opts
    });
  }
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
    // Check Values
    if( p.Name == "" ){
      dialog.showErrorBox("ポーリング編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "savePolling", payload: p }, message => {
      if(message.payload !== "ok") {
        dialog.showErrorBox("ポーリング編集", "保存に失敗しました。");
        return;
      }
      setupPolling();
    });
    pane.dispose();
    pane = undefined;
  });
}

