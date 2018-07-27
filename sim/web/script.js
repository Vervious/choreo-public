
let data = [];
let dataNodeIndex = new Map(); // maps nodeId -> location in data array
let lastRequestedTime = 0;
function initData() {
}
// called by visualization to get the current state of the data
// i is the time at which the function is called
function getData(time) {
    lastRequestedTime = time;
    return data
}

function updateData(edArr) {
    // ed[0] has format ed.NodeID -> int; ed.ClockTime -> int
    edArr.forEach(function(ed) {
        nId = ed.NodeID;
        if (dataNodeIndex.has(nId) === false) {
            dataNodeIndex.set(nId, data.length);
            const datum = {
                NodeID: nId,
                ClockTime: ed.ClockTime,
                LastClockTime: 0,
                LastChangeTime: lastRequestedTime,
                Corrected: ed.Corrected,
            };
            data.push(datum);
        } else {
            i = dataNodeIndex.get(nId);
            data[i].LastClockTime = data[i].ClockTime;
            data[i].ClockTime = ed.ClockTime;
            data[i].LastChangeTime = lastRequestedTime;
            data[i].Corrected = ed.Corrected;
        }
    });
}

function main_data() {

var lastMsg = -1;
var domain = window.location.hostname;
// var domain = "bob.expert"
console.log("connecting to: " + domain);
msgSource = new EventSource(`http://` + domain + `:8910/ev`);
// Attach listeners
msgSource.addEventListener('message', processMessage);
// Render new messages as they arrive
function processMessage(e) {
    let msgId = parseInt(e.lastEventId, 10);
    if (msgId > lastMsg) {
        lastMsg = msgId;
        ed = JSON.parse(e.data);
        // console.log(ed);
        // console.log(`${msgId}:\t[${ed.NodeID}]\t->${ed.ClockTime}`);
        updateData(ed); // ed is an array of updates
    }
}

}


function main(err, regl) {
    

const drawDots = regl({
    // draw circles
    // https://www.desultoryquest.com/blog/drawing-anti-aliased-circular-points-using-opengl-slash-webgl/
  frag: `
  precision mediump float;
  varying vec4 fragColor;
  void main () {
    float r = 0.0;
    vec2 cxy = 2.0 * gl_PointCoord - 1.0;
    r = dot(cxy, cxy);
    if (r > 1.0) {
        discard;
    }
    gl_FragColor = fragColor;
  }`,

  vert: `
  precision mediump float;
  attribute vec2 startPosition;
  attribute vec2 endPosition;
  attribute float timeSinceChange;
  attribute vec4 color;

  uniform float pointWidth;
  uniform float animationDuration;

  varying vec4 fragColor;

  float easeCubicInOut(float t) {
    t *= 2.0;
    t = (t <= 1.0 ? t * t * t : (t -= 2.0) * t * t + 2.0) / 2.0;

    if (t > 1.0) {
      t = 1.0;
    }
    return t;
  }
  void main () {
    float t;
    fragColor = color;
    // vec2 position;
    // if (animationDuration == 0) {
    //     position = endPosition;
    // } else {
    //     t = easeCubicInOut(timeSinceChange * 1000. / animationDuration);
    //     position = mix(startPosition, endPosition, t);
    // }
    gl_PointSize = pointWidth;
    gl_Position = vec4(endPosition, 0, 1);
  }`,

  attributes: {
    endPosition: function(context, props) {
        const coords = (clockTime, id) => {
            let NDC = [clockTime / 100.0 - 1, (id - 50.0) / 50.0 * 0.5];
            return props.projectFn(NDC);
        };
        return props.data.map(ed => coords(ed.ClockTime, ed.NodeID));
    },
    startPosition: function(context, props) {
        const coords = (clockTime, id) => {
            let NDC = [clockTime / 100.0 - 1, (id - 50.0) / 50.0 * 0.5];
            return props.projectFn(NDC);
        };
        return props.data.map(ed => coords(ed.LastClockTime, ed.NodeID));
    },
    timeSinceChange: function(context, props) {
        return props.data.map(ed => (context.time - ed.LastChangeTime));
    },
    color: function(context, props) {
        const color = (clockTime, corrected) => {
            let color2d = clockTime % 600;
            let R = color2d / 600.0;
            let B = 1;
            let G = 0.304;
            if (corrected === 1) {
                R = 0;
                B = 0;
                G = 1;
            }
            return [R, G, B, 1.000];
        }
        return props.data.map(ed => color(ed.ClockTime, ed.Corrected));
    },
  },

  uniforms: {
    animationDuration: 0,
    pointWidth: function(context, props) {
        return 10 * props.devicePixelRatio;
    },
  },

  primitive: "points",

  count:  function(context, props) {
    return props.data.length;
  },
});

const drawMap = regl({

  frag: `
  precision mediump float;
  uniform vec4 color;
  void main () {
    gl_FragColor = color;
  }`,

  vert: `
  precision mediump float;
  attribute vec2 position;
  uniform float pointWidth;
  void main () {
    gl_Position = vec4(position, 0, 1);
  }`,

  attributes: {
    position: (new Array(5)).fill().map((x, i) => {
      var theta = 2.0 * Math.PI * i / 5
      return [ Math.sin(theta), Math.cos(theta) ]
    }),
  },

  uniforms: {
    color: function(context, props) {
        return props.color;
    },
  },

  elements: [
    [0, 1],
    [0, 2],
    [0, 3],
    [0, 4],
    [1, 2],
    [1, 3],
    [1, 4],
    [2, 3],
    [2, 4],
    [3, 4]
  ],

  lineWidth: 2
});



const width = window.innerWidth;
const height = window.innerHeight;
// this is the coordinate space of our display.
// projects a point in NDC to another point in NDC
const projectFn = ([x, y]) => {
    // this basically loops the rectangle on itself into a circle
    const r = (y + 1)/2;
    const theta = (x + 1) * Math.PI;
    return [r * Math.cos(theta), r * Math.sin(theta)];
}

// We use a custom run loop to run our animation at a lower framerate,
// reducing the performance demands of the visualization. For 60Hz use regl.frame, and comment out
// the poll() and timeout lines.
var devicePixelRatio = window.devicePixelRatio || 1;
const loopFn = ({time}) => {
    // regl.poll();

    // regl.clear({
    //     // background color (white)
    //     color: [1, 1, 1, 1],
    // });
    drawDots({
        projectFn: projectFn,
        data: getData(time),
        devicePixelRatio: devicePixelRatio,
    });
    drawMap({
        projectFn: projectFn,
        color: [0.5, 0.5, 0.5, 1.0],
        position: [-1, -1],
    });

    // setTimeout(loopFn, 100, i+1); // in ms
}
// loopFn(0);
regl.frame(loopFn)

// run the main data function
main_data();

}

window.addEventListener('load', function() {
createREGL({onDone: main, container: document.getElementById("regldiv")});
})



