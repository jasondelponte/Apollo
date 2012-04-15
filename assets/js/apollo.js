$(document).ready(function() {
    var conn = null;
    var ctx = null;

    var _apollo = window.apolloApp;

    function drawBlock(block) {
        ctx.fillStyle = "rgba("+block.R+","+block.G+","+block.B+", 0."+block.A+")";
        ctx.fillRect (block.X, block.Y, block.W, block.H);
    }

    if (window["WebSocket"]) {
    	var wsURL = "ws://" + _apollo.host + _apollo.path + "/ws-gn";

        conn = new WebSocket(wsURL);
        conn.onclose = function(evt) {
            console.log('Connection Closed');
        }
        conn.onmessage = function(evt) {
            var block = JSON.parse(evt.data).Block;
            drawBlock(block);
        }
    } else {
        $('.no-websockets').removeClass('hidden');
    }

    var canvas = $('#game-board')[0];
    if (canvas.getContext){
        ctx = canvas.getContext('2d');

    } else {
        $('.no-canvas').removeClass('hidden');
    }
});