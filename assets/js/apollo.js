$(document).ready(function() {
    var conn = null;
    var ctx = null;
    var entities = [];

    var PLAYER_CMD = {
        GAME: {
            REMOVE_ENTITY:0
        }
    };
    var UPDATE_TYPES = {ENTITY_REMOVE: 0, ENTITY_ADD: 1};
    var ENTITY_TYPES = {BLOCK:0};

    var _apollo = window.apolloApp;

    function addBlock(block) {
        entities.push(block);
        ctx.fillStyle = "rgba("+block.R+","+block.G+","+block.B+",0."+block.A+")";
        ctx.fillRect(block.X, block.Y, block.W, block.H);
    }

    function removeBlock(block) {
        ctx.clearRect(block.X, block.Y, block.W, block.H);

        for (var idx=0; idx < entities.length; idx++) {
            if (entities[idx].T === block.T) {
                found = idx;
                break;
            }
        }
        if (idx !== entities.length) {
            entities.splice(idx, 1)
        } else {
            console.log('entitiy not found, '+JSON.stringify(block));
        }
    }

    $('#game-board').click(function() {
        if (!conn || entities.length === 0) {
            console.log('unable to send', conn, entities.length);
            return
        }
        entityRemove = {
            Act: {
                G: {
                    C: PLAYER_CMD.GAME.REMOVE_ENTITY,
                    E: entities[0].ID
                }
            }
        };
        removeBlock(entities[0]);
        conn.send(JSON.stringify(entityRemove));
    });

    if (window["WebSocket"]) {
    	var wsURL = "ws://" + _apollo.wsHost + _apollo.path + "/ws";

        conn = new WebSocket(wsURL);
        conn.onclose = function(evt) {
            console.log('Connection Closed');
            conn = null
        }
        conn.onmessage = function(evt) {
            var msg = JSON.parse(evt.data);
            if (msg.BU) { // Board updates
                var updates = msg.BU;
                for (var idx=0; idx < updates.length; idx++) {
                    var update = updates[idx]

                    if (update.T === UPDATE_TYPES.ENTITY_ADD) { // Update Type, Add
                        if (update.E.T === ENTITY_TYPES.BLOCK) { // Entity Type Block
                            addBlock(update.E);
                        }
                    } else if (update.T === UPDATE_TYPES.ENTITY_REMOVE) { // Update type, Remove
                        if (update.E.T === ENTITY_TYPES.BLOCK) { // Entity type block
                            removeBlock(update.E);
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