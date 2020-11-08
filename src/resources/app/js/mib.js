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
    "info": false,
    "searching": true,
    "autoWidth": true,
    "columns": [],
    "language": {
      "decimal":        "",
      "emptyTable":     "表示するMIBがありません。",
      "info":           "全 _TOTAL_ 件中 _START_ - _END_ 表示",
      "infoEmpty":      "",
      "infoFiltered":   "(全 _MAX_ 件)",
      "infoPostFix":    "",
      "thousands":      ",",
      "lengthMenu":     "_MENU_ 件表示",
      "loadingRecords": "読み込み中...",
      "processing":     "処理中...",
      "search":         "フィルター:",
      "zeroRecords":    "一致するMIBがありません。",
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
    let i = vb.indexOf("=");
    if(i > 0){
      const name = vb.substring(0,i);
      const val  = vb.substring(i+1);
      i = name.indexOf(".");
      if(i > 0){
        const base = name.substring(0,i);
        const index = name.substring(i+1);
        if(!names.includes(base)){
          names.push(base);
        }
        if(!indexes.includes(index)){
          indexes.push(index);
          rows.push([index])
        }
        const r =  indexes.indexOf(index);
        if(r>=0){
          rows[r].push(val);
        }
      }
    }
  });
  const cols = [{title:"Index"}];
  names.forEach(n =>{
    cols.push({title:n})
  })
  makeMibTable(cols);
  mibTable.clear();
  rows.forEach(r => {
    mibTable.row.add(r);
  });
  mibTable.draw();
}

document.addEventListener('astilectron-ready', function () {
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "setParams":
        if (message.payload && message.payload.NodeID) {
          nodeID = message.payload.NodeID;
          mibNames = message.payload.MibNames;
          setWindowTitle(message.payload.NodeName);
          makeMibTable([
            {title:"Index" },
            {title:"名前" },
            {title:"値" },
          ]);
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
  $('.toolbar-actions button.mibtree').click(() => {
    showMibTree();
  });
  $('.toolbar-actions button.reload').click(() => {
    updateMibTree();
  });
  $('.toolbar-actions button.get').click(() => {
    hideMibTree();
    const params = {
      NodeID: nodeID,
      Name: $(".mib_btns input[name=mib]").val()
    }
    $('.toolbar-actions button.get').prop("disabled", true);
    $('#wait').removeClass("hidden");
    astilectron.sendMessage({ name: "get", payload: params }, message => {
      $('#wait').addClass("hidden");
      $('.toolbar-actions button.get').prop("disabled", false);
      const vbl = message.payload;
      if(typeof vbl === "string"){
        setTimeout(() => {
          dialog.showErrorBox("エラー", message.payload);
        }, 100);
        return;
      }
      if(params.Name.indexOf("Table") != -1 ) {
        showTable(vbl)
        return;
      }
      makeMibTable([
        {title:"Index" },
        {title:"名前" },
        {title:"値" },
      ]);    
      mibTable.clear();
      let i = 1;
      vbl.forEach(vb => {
        const a = vb.split('=',2)
        if (a.length > 1){
          mibTable.row.add([i,a[0],a[1]]);
          i++;
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
    minLength: 2
  },
  {
    name: 'MibName',
    limit: 200,
    source: mn()
  });  
});

function setWindowTitle(n){
  const t = "MIBブラウザー - " + n;
  $("title").html(t);
  $("h1.title").html(t);
}

let mibTree;

function showMibTree() {
  if(!mibTree) {
    makeMibTree();
    updateMibTree();
  }
  $("#mibtree_page").toggleClass("hidden");
  $(".toolbar-footer button.reload").toggleClass("hidden");
  $("#mib_page").toggleClass("hidden");
}

function hideMibTree() {
  $("#mibtree_page").addClass("hidden");
  $(".toolbar-footer button.reload").addClass("hidden");
  $("#mib_page").removeClass("hidden");
}

function makeMibTree() {
  const option = {
    tooltip: {
      trigger: 'item',
      triggerOn: 'mousemove',
      formatter: '{c}'
    },
    series: [
      {
        type: 'tree',
        id: 0,
        name: 'mibtree',
        data: [],
        top: '1%',
        left: '10%',
        bottom: '1%',
        right: '10%',

        symbolSize: 6,

        edgeShape: 'polyline',
        edgeForkPosition: '50%',
        initialTreeDepth: 3,

        lineStyle: {
          width: 2
        },

        label: {
          backgroundColor: '#fff',
          position: 'left',
          verticalAlign: 'middle',
          align: 'right'
        },

        leaves: {
          label: {
            position: 'right',
            verticalAlign: 'middle',
            align: 'left'
          }
        },
        expandAndCollapse: true,
        animationDuration: 550,
        animationDurationUpdate: 750
      }
    ]
  };
  mibTree = echarts.init(document.getElementById('mibtree'));
  mibTree.setOption(option);
  mibTree.on('click', params => {
    $(".mib_btns input[name=mib]").val(params.data.name);
  }); 
}

function updateMibTree() {
  astilectron.sendMessage({ name: "mibtree", payload: "" }, message => {
    const js = message.payload;
    if(typeof js != "string" ){
      setTimeout(() => {
        dialog.showErrorBox("エラー","MIBツリー取得エラー" );
      }, 100);
      return;
    }
    const data  = JSON.parse(js);
    const option = {
      series: [{data: [data],}]
    };
    mibTree.setOption(option);
  });
}