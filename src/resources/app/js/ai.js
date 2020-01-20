'use strict';

let epoch = 20;
console.log(tf.getBackend());

document.addEventListener('astilectron-ready', function () {
  astilectron.onMessage(function (message) {
    switch (message.name) {
      case "doAI":
        if( !message.payload || !message.payload.PollingID ){
          return { name: "doAI", payload: "ng" };
        }
        setTimeout(() => {
          doAI(message.payload);
        },100);
        return { name: "doAI", payload: "ok" };
      case "deleteModel":
        deleteModel(message.payload);
        return { name: "deleteModel", payload: "ok" };
      case "clearAllAIMoldes":
        clearAllAIMoldes();
        console.log("clearAllAIMoldes");
        return { name: "clearAllAIMoldes", payload: "ok" };
      case "error":
        setTimeout(() => {
          astilectron.showErrorBox("エラー", message.payload);
        }, 100);
        return { name: "error", payload: "ok" };
    }
  });
});

function deleteModel(PollingID){
  const modelPath = `indexeddb://twsnmpai-${PollingID}`;
  try {
    tf.io.removeModel(modelPath);
    console.log("deleteModel "+modelPath)
  } catch (e) {
    console.log(e);
  }
}

function clearAllAIMoldes(){
  try {
    tf.io.listModels().then(models=>{
      for(let url in models){
        console.log("clearAllAIMoldes "+ url)
        tf.io.removeModel(url);
      }
    });
  } catch (e) {
    console.log(e);
  }
}

async function doAI(req) {
  const modelPath = `indexeddb://twsnmpai-${req.PollingID}`;
  const dataLen = req.Data[0].length
  let autoencoder;
  try {
    autoencoder = await tf.loadLayersModel(modelPath);
  } catch (e) {
    console.log(e);
    const input = tf.input({ shape: [dataLen] });
    const encoded1 = tf.layers.dense({ units: Math.ceil(dataLen/2), activation: 'relu' });
    const encoded2 = tf.layers.dense({ units: Math.ceil(dataLen/4), activation: 'relu' });
    const encoded3 = tf.layers.dense({ units: Math.ceil(dataLen/2), activation: 'relu' });
    const decoded = tf.layers.dense({ units: dataLen, activation: 'sigmoid' });
    const output = decoded.apply(encoded3.apply(encoded2.apply(encoded1.apply(input))));
    autoencoder = tf.model({ inputs: input, outputs: output });
  }
  autoencoder.compile({ optimizer: 'adam', loss: 'meanSquaredError' });
  
  const lossData = [];
  const x_train = tf.tensor2d(req.Data, [req.Data.length, dataLen]);
  for (let i = 0; i < epoch; i++) {
    const h = await autoencoder.fit(x_train, x_train, { epochs: 5, batchSize: 24 });
    lossData.push([moment().valueOf(), h.history.loss[0]]);
    console.log("Loss after Epoch " + i + " : " + h.history.loss[0]);
  }
  // Saveできない場合も終了するため
  try {
    await autoencoder.save(modelPath);
  } catch(e){
    console.log(e);
  }
  const evd = [];
  for (let i = 0; i < req.Data.length; i++) {
    const x_eval = tf.tensor2d(req.Data[i], [1, dataLen]);
    const r = await autoencoder.evaluate(x_eval, x_eval, {});
    evd.push(r.dataSync()[0]);
  }
  let avg = average(evd);
  let sd = standardDeviation(evd, avg);
  let ssArr = standardScore(evd, avg, sd);
  let scoreData = [];
  for (let i = 0; i < req.Data.length; i++) {
    scoreData.push([req.TimeStamp[i],ssArr[i]])
  }
  astilectron.sendMessage({ name: "done", payload:{
    PollingID:req.PollingID,
    LastTime: req.TimeStamp[req.Data.length-1],
    LossData: lossData,
    ScoreData:scoreData
    }}, message => {
      console.log(message);
  });
}

function average(x) {
  let n = x.length;
  let avg = 0;
  for (let i = 0; i < n; i++) {
    avg += x[i];
  }
  return avg / n; //  (1 / n * avg)
}

function standardDeviation(x, avg) {
  let n = x.length;
  let sum = 0;
  for (let i = 0; i < n; i++) {
    sum += Math.pow(x[i] - avg, 2);
  }
  return Math.sqrt(sum / n);
}

function standardScore(x, avg, sd) {
  let ssArr = [];
  let n = x.length;
  for (let i = 0; i < n; i++) {
    let ti = Math.round((10 * (x[i] - avg) / sd) + 50);
    // if (ti > 100) {
    //   ti = 100;
    // }
    ssArr.push(ti);
  }
  return ssArr;
}
