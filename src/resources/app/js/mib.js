'use strict';

let nodeID = "";
let mibNames;
let mibTable;


function makeMibTable(cols) {
  if(mibTable) {
    mibTable.destroy();
    const table = `<table id="mib_table" class="display compact">
    <thead>
    </thead>
    <tbody>
    </tbody>
  </table>`;
    $("div.table_base").html(table);
  }
  const opt =  {
    "paging": true,
    "info": false,
    "searching": true,
    "autoWidth": true,
    "columns": [],
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
      "search":         "検索:",
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
  }
  opt.columns = cols;
  mibTable = $('#mib_table').DataTable(opt);
}

function showTable(vbl) {
  const names = [];
  const indexes = [];
  const rows = [];
  vbl.forEach(vb=>{
    const a = vb.split("=",2)
    if(a.length == 2){
      const b = a[0].split(".",2)
      if(b.length == 2 ){
        if(!names.includes(b[0])){
          names.push(b[0]);
        }
        if(!indexes.includes(b[1])){
          indexes.push(b[1]);
          rows.push([b[1]])
        }
        const r =  indexes.indexOf(b[1]);
        if(r>=0){
          rows[r].push(a[1]);
        }
      }
    }
  });
  const cols = [{title:"Index"}];
  names.forEach(n =>{
    cols.push({title:n})
  })
  makeMibTable(cols);
  mibTable.rows().remove();
  rows.forEach(r => {
    mibTable.row.add(r);
  });
  mibTable.draw();
}

document.addEventListener('astilectron-ready', function () {
  makeMibTable([
    {title:"名前" },
    {title:"値" },
  ]);
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "setParams":
        if (message.payload && message.payload.NodeID) {
          nodeID = message.payload.NodeID;
          mibNames = message.payload.MibNames;
          setWindowTitle(message.payload.NodeName);
        }
        return { name: "setNodeID", payload: "ok" };
      case "error":
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
  $('.toolbar-actions button.get').click(() => {
    const params = {
      NodeID: nodeID,
      Name: $(".mib_btns input[name=mib]").val()
    }
    astilectron.sendMessage({ name: "get", payload: params }, message => {
      const vbl = message.payload;
      if(params.Name.indexOf("Table") != -1 ) {
        showTable(vbl)
        return;
      }
      makeMibTable([
        {title:"名前" },
        {title:"値" },
      ]);    
      mibTable.rows().remove();
      vbl.forEach(vb => {
        const a = vb.split('=',2)
        if (a.length > 1){
          mibTable.row.add(a);
        } 
      });
      mibTable.draw();
    });
  });
  const mn = function() {
    return function findMatches(q, cb) {
      let matches, substrRegex;  
      // an array that will be populated with substring matches
      matches = [];
      // regex used to determine if a string contains the substring `q`
      substrRegex = new RegExp(q, 'i');
      // iterate through the pool of strings and for any string that
      // contains the substring `q`, add it to the `matches` array
      $.each(mibNames, function(i, str) {
        if (substrRegex.test(str)) {
          matches.push(str);
        }
      });  
      cb(matches);
    };
  };
  $('.mib_btns input[name=mib]').typeahead({
    hint: true,
    highlight: true,
    minLength: 1
  },
  {
    name: 'MibName',
    source: mn()
  });  
});

function setWindowTitle(n){
  const t = "MIBブラウザー - " + n;
  $("title").html(t);
  $("h1.title").html(t);
}
