'use strict';

let polling;
let node;
let logTable;
let logChart;
let stateChart;
let resultChart;
let currentPage;

function showPage(mode) {
  const pages = ["log", "state", "result"];
  pages.forEach(p => {
    if (mode == p) {
      $("#" + p + "_page").removeClass("hidden");
      $("#" + p).addClass("active");
    } else {
      $("#" + p + "_page").addClass("hidden");
      $("#" + p).removeClass("active");
    }
  });
  currentPage = mode;
}

function makeLogTable() {
  const opt =  {
    "order": [[1, "desc"]],
    "paging": true,
    "info": false,
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
  logTable = $('#log_table').DataTable(opt);
}

function makeLogChart() {
  const option = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      formatter: function (params) {
        const p = params[0];
        return p.name + ' : ' + p.value[1];
      },
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: "5%",
      right:"5%",
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      axisLabel:{
        fontSize: "8px",
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('MM/dd hh:mm', date)
        }
      },
      splitLine: {
        show: false
      },
    },
    yAxis: {
      type: 'value',
    },
    series: [{
      type: 'bar',
      color: "#1f78b4",
      large: true,
      data: [],
    }]
  };
  logChart = echarts.init(document.getElementById('log_chart'));
  logChart.setOption(option);
}

function  makeStateChart(){
  const option = {
      tooltip : {
          trigger: 'axis',
          axisPointer : {
              type : 'shadow'
          }
      },
      color:[ "#e31a1c","#fb9a99","#dfdf22","#33a02c","#999"],
      legend: {
          data: ['重度','軽度','注意','正常','不明']
      },
      grid: {
          left: '7%',
          right: '4%',
          bottom: '3%',
          containLabel: true
      },
      xAxis:  {
          type: 'value'
      },
      yAxis: {
          type: 'category',
          data: []
      },
      series: [
        {
          name: '重度',
          type: 'bar',
          stack: '回数',
          data: []
        },
        {
          name: '軽度',
          type: 'bar',
          stack: '回数',
          data: []
        },
        {
          name: '注意',
          type: 'bar',
          stack: '回数',
          data: []
        },
        {
          name: '正常',
          type: 'bar',
          stack: '回数',
          data: []
        },
        {
          name: '不明',
          type: 'bar',
          stack: '回数',
          data: []
        },
      ],
  };
  stateChart = echarts.init(document.getElementById('state_chart'));
  stateChart.setOption(option);
}

function makeResultChart() {
  const option = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      formatter: function (params) {
        const p = params[0];
        return p.name + ' : ' + p.value[1];
      },
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: "10%",
      right:"5%",
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      axisLabel:{
        fontSize: "8px",
        formatter: function (value, index) {
          var date = new Date(value);
          return echarts.format.formatTime('MM/dd hh:mm', date)
        }
      },
      splitLine: {
        show: false
      },
    },
    yAxis: {
      type: 'value',
    },
    series: [{
      color: "#1f78b4",
      type: 'line',
      showSymbol: false,
      hoverAnimation: false,
      data: [],
    }]
  };
  resultChart = echarts.init(document.getElementById('result_chart'));
  resultChart.setOption(option);
}


function setupTimeVal() {
  $(".toolbar-actions input[name=start]").val(moment().subtract(1, "h").format("Y-MM-DDTHH:00"));
  $(".toolbar-actions input[name=end]").val(moment().format("Y-MM-DDTHH:00"));
}

function clearData() {
  logTable.rows().remove();
  logTable.draw();
  logChart.setOption({
    series: [{
      data: []
    }]
  });
  logChart.resize();
  const optState = {
    yAxis:{
      data:[],
    },
    series:[
      {data:[]},
      {data:[]},
      {data:[]},
      {data:[]},
      {data:[]},
    ]
  };
  stateChart.setOption( optState);
  stateChart.resize();
  resultChart.setOption({
    series: [{
      data: []
    }]
  });
  resultChart.resize();

}

document.addEventListener('astilectron-ready', function () {
  showPage("log");
  setupTimeVal();
  makeLogTable();
  makeLogChart();
  makeStateChart();
  makeResultChart();
  logChart.resize();
  stateChart.resize();
  resultChart.resize();
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "setParams":
        if (message.payload && message.payload.Polling) {
          polling = message.payload.Polling;
          node = message.payload.Node;
          setWindowTitle(node.Name,polling.Name);
          clearData();
          showPage("log");
          logChart.resize();      
        }
        return { name: "setParams", payload: "ok" };
      case "error":
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
  $('#log').click(()=>{
    showPage("log");
    logChart.resize();
  });
  $('#state').click(()=>{
    showPage("state");
    stateChart.resize();
  });
  $('#chart').click(()=>{
    showPage("result");
    resultChart.resize();
  });
  $('.toolbar-actions button.close').click(() => {
    astilectron.sendMessage({ name: "close", payload: "" }, message => {
    });
  });
  $('.toolbar-actions button.get').click(() => {
    const params = {
      PollingID: polling.ID,
      StartTime: $(".toolbar-actions input[name=start]").val(),
      EndTime:   $(".toolbar-actions input[name=end]").val()
    }
    $('.toolbar-actions button.get').prop("disabled", true);
    astilectron.sendMessage({ name: "get", payload: params }, message => {
      $('.toolbar-actions button.get').prop("disabled", false);
      const logs = message.payload;
      if(typeof logs === "string"){
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return;
      }
      const dataTimeLine = [];
      const dataResult = [];
      const optState = {
        yAxis:{
          data:[],
        },
        series:[
          {data:[]},
          {data:[]},
          {data:[]},
          {data:[]},
          {data:[]},
        ]
      };
      const stateCount = {
        high: 0,
        low: 0,
        warn: 0,
        normal:0,
        unkown:0,
      }
      let intCount = 5;
      let intState = 1;
      if(logs.length > 2){
        const  dh = (logs[logs.length-1].Time - logs[0].Time) / (1000*1000*1000*3600);
        intState = Math.round(dh/4);
        if(intState < 1){
          intState = 1; 
        }
        if(intState > 24) {
          intCount = 60;
        }
      }
      let count = 0;
      let ctc;
      let cts;
      logTable.rows().remove();
      logs.forEach(l => {
        const ts = moment(l.Time / (1000 * 1000)).format("Y/MM/DD HH:mm:ss.SSS");
        const state = getStateHtml(l.State)
        logTable.row.add([state, ts, l.NumVal, l.StrVal]);
        dataResult.push({
          name: ts,
          value:[new Date(l.Time/(1000*1000)),l.NumVal],
        });
        const newCtc = Math.floor(l.Time / (1000 * 1000 * 1000 * 60 * intCount));
        if(!ctc) {
          ctc = newCtc;
        }
        if (ctc != newCtc) {
          let t = new Date(ctc * 60 * intCount * 1000);
          dataTimeLine.push({
            name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
            value: [t,count]
          });
          ctc = newCtc;
          count = 0;
        }
        count++;
        const newCts = Math.floor(l.Time / (1000 * 1000 * 1000 * 60 * 60 * intState));
        if(!cts) {
          cts = newCts;
        }
        if (cts != newCts) {
          let t = new Date(cts * 60 * 60 * 1000 * intState);
          optState.yAxis.data.push(echarts.format.formatTime('MM/dd hh:mm', t));
          optState.series[0].data.push(stateCount.high);
          optState.series[1].data.push(stateCount.low);
          optState.series[2].data.push(stateCount.warn);
          optState.series[3].data.push(stateCount.normal);
          optState.series[4].data.push(stateCount.unkown);
          cts = newCts
          stateCount.high=0;
          stateCount.low=0;
          stateCount.warn=0;
          stateCount.normal=0;
          stateCount.unkown=0;
        }
        switch (l.State){
          case "high":
            stateCount.high++;
            break;
          case "low":
            stateCount.low++;
            break;
          case "warn":
            stateCount.warn++;
            break;
          case "normal":
          case "repair":
            stateCount.normal++;
            break;
          default:
            stateCount.unkown++;
            break;
        }
      });
      if(count > 0 ){
        let t = new Date(ctc * 60 * intCount * 1000);
        dataTimeLine.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
          value: [t,count]
        });
        t = new Date(cts * 60 * 60 * 1000 * intState);
        optState.yAxis.data.push(echarts.format.formatTime('MM/dd hh:mm', t));
        optState.series[0].data.push(stateCount.high);
        optState.series[1].data.push(stateCount.low);
        optState.series[2].data.push(stateCount.warn);
        optState.series[3].data.push(stateCount.normal);
        optState.series[4].data.push(stateCount.unkown);
      }
      logTable.draw();
      logChart.setOption({
        series: [{
          data: dataTimeLine
        }]
      });
      logChart.resize();
      stateChart.setOption( optState);
      stateChart.resize();
      resultChart.setOption({
        series: [{
          data: dataResult
        }]
      });
      resultChart.resize();
    });
  });
});

function setWindowTitle(n,p){
  const t = "ポーリング分析 - " + n +" - " + p;
  $("title").html(t);
  $("h1.title").html(t);
}
