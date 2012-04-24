
var ApolloApp = (function(context){
    var config = null;
    var ws = null;
    var board = null;

    function selected(id) {
        entityRemove = {
            Act: {
                G: {
                    C: WsConn.PlayerGameCmd.selectEntity,
                    E: parseInt(id)
                }
            }
        };
        ws.conn.send(JSON.stringify(entityRemove));
    }


    function runApp(cfg) {
        // Canvas 2d must be supported before we can run
        if (!window.CanvasRenderingContext2D) {
            cfg.noCanvas();
            return
        }
        board = newGameBoard(cfg.container, cfg.width, cfg.height)

        ws = new WsConn(board);
        if (!ws.open(cfg.wsURL)) {
            cfg.noWebSockets()
            return
        }
    }


    // Webseocket wrapper object
    function WsConn(board) {
        this.conn = null;
        this.board = board
    };
    WsConn.PlayerGameCmd = {selectEntity: 0};
    WsConn.EntityUpdateTypes = {added: 0, present: 1, selected: 2, removed: 3};
    WsConn.PlayerUpdateTypes = {added: 0, present: 1, updated: 2, removed: 3};
    WsConn.EntityTypes = {block:0};
    WsConn.prototype.open = function(url) {
        if (window["WebSocket"]) {
            var conn = this.conn = new WebSocket(url);
            var ws = this;
            conn.onclose = function(evt) { ws.onClose(evt); };
            conn.onmessage = function(evt) { ws.onMessage(evt); };
            return true;
        }
        return false
    };
    WsConn.prototype.onClose = function(evt) {
        console.log('Connection Closed,', evt);
        this.conn = null
    };
    WsConn.prototype.onMessage = function(evt) {
        // console.log(evt.data);
        var msg = JSON.parse(evt.data);
        if (msg.GU) { // Game board update
            var entities = msg.Es;
            if (entities) {
                this.processEntityUpdate(entities)
            }
            var players = msg.Ps;
            if (players) {
                this.processPlayerUpdate(players)
            }
        }
    };
    WsConn.prototype.processEntityUpdate = function(entities) {
        var eLen = entities.length;
        for (var idx=0; idx < eLen; idx++) {
            var entity = entities[idx];
            switch(entity.St) {
                case WsConn.EntityUpdateTypes.added:
                    this.board.addEntity(entity);
                break;

                case WsConn.EntityUpdateTypes.present:
                    this.board.addEntity(entity);
                break;
                
                case WsConn.EntityUpdateTypes.selected:
                    this.board.updateEntity(entity);
                break;

                case WsConn.EntityUpdateTypes.removed:
                    this.board.removeEntity(entity.Id);
                break;
            }
        }
    }
    WsConn.prototype.processPlayerUpdate = function(players) {
        var pLen = players.length;
        for (var idx=0; idx < pLen; idx++) {
            var player = players[idx];
            switch (player.St) {
                case WsConn.PlayerUpdateTypes.added:
                    this.board.addPlayer(player);
                break;

                case WsConn.PlayerUpdateTypes.present:
                    this.board.addPlayer(player);
                break;

                case WsConn.PlayerUpdateTypes.updated:
                    this.board.updatePlayer(player)
                break;

                case WsConn.PlayerUpdateTypes.removed:
                    this.board.removePlayer(player.Id)
                break;
            }
        }
    }



    function newGameBoard(container, width, height) {
        var entityColors = ['red', 'blue', 'green', 'gray', 'orange'];
        var stage = null,
            playerLayer = null,
            msgLayer = null,
            entLayer = null,
            players  = [],
            entities = [];

        function writeMsg(layer, msg) {
            var context = layer.getContext();
            layer.clear();
            context.font = "18pt Calibri";
            context.fillStyle = "black";
            context.fillText(msg, 10, 25);
        }

        // Players
        function findPlayer(id) {
            var pLen =  players.length;
            var idx;
            for (idx=pLen-1; idx >= 0; idx--) {
                var player = players[idx];
                if (player.Id === id) {
                    break;
                }
            }
            return idx;
        }
        function addPlayer(player) {
            var idx = findPlayer(player.Id);
            if (idx !== -1) { return; }
            players.push(player);
            playersAdded = true;
        }
        function removePlayer(id) {
            var idx = findPlayer(id);
            if (idx !== -1) {
                players.splice(idx, 1);
            } else {
                console.error('tried to removed a player I dont know about;', player);
            }
        }
        function updatePlayer(player) {
            var idx = findPlayer(player.Id);
            if (idx !== -1) {
                players[idx] = player;
            } else {
                console.error('tried to update a player I dont know about;', player);
            }
        }

        // Entities
        function addEntity(entity) {
            // Inore duplicate items
            if (entities[entity.Id]) { return; }

            entity.color = entityColors[entity.C];
            var rect = new Kinetic.Rect({
                x: entity.X,
                y: entity.Y,
                width:  entity.W,
                height: entity.H,
                fill:   entity.color,
                stroke: "black",
                strokeWidth: 2
            });

            entities[entity.Id] = {e: entity, d: rect};
            rect.on('click', function(evt) {
                selected(entity.Id);
            });
            entLayer.add(rect)
            entLayer.draw();
        }
        function updateEntity(entity) {
            writeMsg(msgLayer, 'entity '+entity.Id+' selected');
            entities[entity.Id].e = entity;
        }
        function removeEntity(id) {
            if (entities[id] && entities[id].d) {
                d = entities[id].d;
                d.clearData();
                d.off('click');
                entLayer.remove(d);
            }
            delete entities[id];
            entLayer.draw();
        }

        function init() {
            stage = new Kinetic.Stage({
                container: container,
                width: width, height: height
            });

            playerLayer = new Kinetic.Layer();
            msgLayer = new Kinetic.Layer();
            entLayer = new Kinetic.Layer();

            stage.add(entLayer);
            stage.add(playerLayer);
            stage.add(msgLayer);
        }


        // Initialize the game board
        init();


        return {
            addPlayer:    addPlayer,
            removePlayer: removePlayer,
            updatePlayer: updatePlayer,
            addEntity:    addEntity,
            updateEntity: updateEntity,
            removeEntity: removeEntity,
        }
    }

    return {
        runApp: runApp,
    };
})(this);
