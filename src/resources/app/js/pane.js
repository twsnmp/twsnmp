'use strict';

function createMapConfPane() {
  const mapConfTmp = mapConf
  const pane = new Tweakpane();

  const f1 = pane.addFolder({
    title: 'マップ設定',
  });
  f1.addInput(mapConfTmp, 'MapName', { label: "名前" });
  f1.addInput(mapConfTmp, 'BackImg', { label: "背景画像" });
  f1.addButton({
    title: '背景画像ファイル選択',
  }).on('click', (value) => {
    astilectron.showOpenDialog({ properties: ['openFile'], title: "背景画像ファイル" }, function (paths) {
      mapConfTmp.BackImg = paths[0];
      pane.refresh();
    });
  }); 
 
  const f2 = pane.addFolder({
    title: 'ポーリング',
  });
  f2.addInput(mapConfTmp, 'PollInt', { 
    label: "間隔",
    min: 60,
    max: 600,
    step: 10,
  });
  f2.addInput(mapConfTmp, 'Timeout', { 
    label: "Timeout",
    min: 1,
    max: 5,
    step: 1,
  });
  f2.addInput(mapConfTmp, 'Retry', { 
    label: "Retry",
    min: 0,
    max: 5,
    step: 1,
  });
  f2.addInput(mapConfTmp, 'Community', { label: "Community" });
  
  const f3 = pane.addFolder({
    title: '受信',
  });
  f3.addInput(mapConfTmp, 'EnableSyslogd', { 
    label: "Syslog",
    options: {
      "Enable": true,
      "Disable": false,
    },
  });
  f3.addInput(mapConfTmp, 'EnableTrapd', { 
    label: "SNMP Trap",
    options: {
      "Enable": true,
      "Disable": false,
    },
  });
  f3.addInput(mapConfTmp, 'EnableNetflowd', { 
    label: "Netflow",
    options: {
      "Enable": true,
      "Disable": false,
    },
  });
  const f4 = pane.addFolder({
    title: 'ログ',
  });
  f4.addInput(mapConfTmp, 'LogDispSize', { 
    label: "表示件数",
    min:1000,
    max:20000,
    step:1000,
 });
  f4.addInput(mapConfTmp, 'LogDays', { 
    label: "保存日数",
    min:1,
    max:365,
    step:1,
  });

  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
  });
  pane.addButton({
    title: 'Save',
  }).on('click', (value) => {
    // Check Values
    if( mapConfTmp.MapName == "" ){
      astilectron.showErrorBox("マップ設定", "マップ名を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "mapConf", payload: mapConf }, message => {
      if(message.payload !== "ok") {
        astilectron.showErrorBox("マップ設定", "保存に失敗しました。");
        return;
      }
      mapConf = mapConfTmp;
    });
    pane.dispose();
  });
  return;
}

function createStartDiscoverPane(x,y) {
  astilectron.sendMessage({ name: "getDiscoverConf", payload: "" }, message => {
    if(!message.payload) {
      astilectron.showErrorBox("自動発見", "設定を取得できません。");
      return;
    }
    const discoverConf= message.payload;
    discoverConf.X= x;
    discoverConf.Y= y;
    const pane = new Tweakpane({
      title: "自動発見"
    });
    pane.addInput(discoverConf, 'StartIP', { label: "開始IP" });
    pane.addInput(discoverConf, 'EndIP', { label: "終了IP" });
    pane.addInput(discoverConf, 'Timeout', { 
      label: "Timeout",
      min: 1,
      max: 5,
      step: 1,
    });
    pane.addInput(discoverConf, 'Retry', { 
      label: "Retry",
      min: 0,
      max: 5,
      step: 1,
    });
    pane.addInput(discoverConf, 'Community', { label: "Community" });
    pane.addButton({
      title: 'Cancel',
    }).on('click', (value) => {
      pane.dispose();
    });
    pane.addButton({
      title: 'Start',
    }).on('click', (value) => {
      // Check Values
      if (discoverConf.StartIP === "" || discoverConf.EndIP === ""  ) {
        astilectron.showErrorBox("範囲指定エラー", "開始、終了IPアドレスが正しくありません。")
        return;
      }
      astilectron.sendMessage({ name: "startDiscover", payload: discoverConf }, message => {
        if(message.payload !== "ok") {
          astilectron.showErrorBox("自動発見", "開始できません。");
          return;
        }
      });
      pane.dispose();
    });  
  });
}

function createEditNodePane(x,y,nodeID) {
  let node;
  if(nodeID != "") {
    node = nodes[nodeID];
  } else {
    node = {
      ID: "",
      Name: "",
      Icon: "desktop",
      Descr: "",
      X: x,
      Y: y,
      IP: "",
      Community: "",
    };
  }
  const pane = new Tweakpane({
    title: nodeID === "" ? "新規ノード" : "ノード編集"
  });
  pane.addInput(node, 'Name', { label: "名前" });
  pane.addInput(node, 'IP', { label: "IPアドレス" });
  pane.addInput(node, 'Icon', { 
    label: "アイコン",
    options: {
      "デスクトップPC": "desktop",
      "ノートPC": "laptop",
      "タブレット" : "tablet",
      "モバイル": "mobile-alt",
      "サーバー": "server",
      "ルーター": "sync",
      "ネットワーク機器": "hdd",
      "プリンター": "print",
      "有線LAN": "network-wired",
      "無線LAN": "wifi",
      "クラウド": "cloud",
    },
  });
  pane.addInput(node, 'Community', { label: "Community" });
  pane.addInput(node, 'Descr', { label: "説明" });
  pane.addButton({
    title: 'Cancel',
  }).on('click', (value) => {
    pane.dispose();
  });
  pane.addButton({
    title: 'Save',
  }).on('click', (value) => {
    // Check Values
    if( node.Name == "" ){
      astilectron.showErrorBox("ノード編集", "名前を指定してください。");
      return;
    }
    astilectron.sendMessage({ name: "saveNode", payload: node }, message => {
      if(message.payload !== "ok") {
        astilectron.showErrorBox("ノード編集", "保存に失敗しました。");
        return;
      }
    });
    pane.dispose();
  });
}

function createEditLinePane(nodeID1,nodeID2) {
  astilectron.sendMessage({ name: "getLine", payload: {NodeID1:nodeID1,NodeID2:nodeID2} }, message => {
    if(!message.payload) {
      astilectron.showErrorBox("ライン編集", "ライン情報を取得できません。");
      return;
    }
    const lineDlg  = message.payload;
    const line = lineDlg.Line;
    const pane = new Tweakpane({
      title: lineDlg.Line.ID  === "" ? "新規ライン" : "ライン編集"
    });
    const n1 = pane.addFolder({
      title: lineDlg.NodeName1,
    });
    const opt1 = {};
    lineDlg.Pollings1.forEach( p => {
      opt1[p.Name] = p.ID;
    });
    n1.addInput(lineDlg.Line, 'PollingID1', { 
      label: "ポーリング",
      options: opt1,
    });
    const n2 = pane.addFolder({
      title: lineDlg.NodeName2,
    });
    const opt2 = {};
    lineDlg.Pollings2.forEach( p => {
      opt2[p.Name] = p.ID;
    });
    n2.addInput(line, 'PollingID2', { 
      label: "ポーリング",
      options: opt2,
    });
    pane.addButton({
      title: 'Cancel',
    }).on('click', (value) => {
      pane.dispose();
    });
    if( line.ID != "" ){
      pane.addButton({
        title: 'Delete',
      }).on('click', (value) => {
        astilectron.sendMessage({ name: "deleteLine", payload: line }, message => {
          if(message.payload !== "ok") {
            astilectron.showErrorBox("ライン編集", "削除に失敗しました。");
            return;
          }
        });
        pane.dispose();
      });
    }
    pane.addButton({
      title: 'Save',
    }).on('click', (value) => {
      // Check Values
      if( line.PollingID1 === "" || line.PollingID1 === ""  ){
        astilectron.showErrorBox("ライン編集", "ポーリングを指定してください。");
        return;
      }
      astilectron.sendMessage({ name: "saveLine", payload: line }, message => {
        if(message.payload !== "ok") {
          astilectron.showErrorBox("ライン編集", "保存に失敗しました。");
          return;
        }
      });
      pane.dispose();
    });
  });
}
