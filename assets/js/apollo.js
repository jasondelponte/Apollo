$(document).ready(function() {
    var conn = null;
    var ctx = null;

    var _apollo = window.apolloApp;

    function addBlock(block) {
        ctx.fillStyle = "rgba("+block.R+","+block.G+","+block.B+",0."+block.A+")";
        ctx.fillRect(block.X, block.Y, block.W, block.H);
    }

    if (window["WebSocket"]) {
    	var wsURL = "ws://" + _apollo.wsHost + _apollo.path + "/ws";

        conn = new WebSocket(wsURL);
        conn.onclose = function(evt) {
            console.log('Connection Closed');
        }
        conn.onmessage = function(evt) {
            var msg = JSON.parse(evt.data);
            if (msg.BU) { // Board updates
                var updates = msg.BU;
                for (var idx=0; idx < updates.length; idx++) {
                    var update = updates[idx]

                    if (update.T === 1) { // Update Type, Add
                        if (update.E.T === 0) { // Entity Type Block
                            addBlock(update.E);
                        }
                    }
                }
            }
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