
var ApolloApp = (function(context){
    var config = null;
    var ws = null;
    var render = null;

    function playerTouch(evt) {
        var id = null;
        for (id in render.entities) {
            if (!render.entities.hasOwnProperty(id)) { continue; }
            break;
        }
        if (id === null) { return; }
        entityRemove = {
            Act: {
                G: {
                    C: WsConn.PlayerCmd.GameRemoveEntity,
                    E: parseInt(id)
                }
            }
        };
        ws.conn.send(JSON.stringify(entityRemove));
    }

    function runApp(cfg) {
        config = cfg
        config.fps = 10;

        render = new Render(config.canvas[0]);
        if (!render.init()) {
            config.noCanvas();
            return
        }

        ws = new WsConn();
        if (!ws.open(config.wsURL)) {
            config.noWebSockets()
            return
        }
        
        render.start(config.fps);

        // TODO really?...
        if ('ontouchstart' in window) {
            config.canvas.bind('touchstart', playerTouch); // Touch, https://dvcs.w3.org/hg/webevents/raw-file/tip/touchevents.html
        } else {
            config.canvas.bind('click', playerTouch); // Mouse
        }
    }


    // Webseocket wrapper object
    function WsConn() {
        this.conn = null;
    };
    WsConn.PlayerCmd = {
        GameRemoveEntity: 0
    };
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
                    render.addEntity(entity);
                break;

                case WsConn.EntityUpdateTypes.removed:
                    render.removeEntityById(entity.Id);
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
                    render.addPlayer(player);
                break;

                case WsConn.PlayerUpdateTypes.present:
                    render.addPlayer(player);
                break;

                case WsConn.PlayerUpdateTypes.updated:
                    render.updatePlayer(player)
                break;

                case WsConn.PlayerUpdateTypes.removed:
                    render.removePlayer(player)
                break;
            }
        }
    }


    // Rendering for drawing 
    function Render(canvas) {
        this.canvas = canvas;
        this.ctx = null;
        this.timer = null;
        this.entities = {};
        this.changed = false;
        this.players = [];
        this.playersAdded = false;
    };
    function playerSort (a, b) {
        return a.Id - b.Id;
    };
    Render.prototype.init = function() {
        if (this.canvas.getContext){
            this.ctx = this.canvas.getContext('2d');
            return true
        }
        return false
    };
    Render.prototype.addPlayer = function(player) {
        var idx = this.findPlayer(player.Id);
        if (idx !== -1) { return; }
        this.players.push(player);
        this.playersAdded = true;
    }
    Render.prototype.removePlayer = function(player) {
        var idx = this.findPlayer(player.Id);
        if (idx !== -1) {
            this.players.splice(idx, 1);
        } else {
            console.error('tried to removed a player I dont know about;', player);
        }
    }
    Render.prototype.updatePlayer = function(player) {
        var idx = this.findPlayer(player.Id);
        if (idx !== -1) {
            this.players[idx] = player;
        } else {
            console.error('tried to update a player I dont know about;', player);
        }
    }
    Render.prototype.findPlayer = function(pId) {
        var pLen =  this.players.length;
        var idx;
        for (idx=pLen-1; idx >= 0; idx--) {
            var player = this.players[idx];
            if (player.Id === pId) {
                break;
            }
        }
        return idx;
    }
    Render.prototype.addEntity = function(e) {
        this.entities[e.Id] = e;
        this.changed = true;
    };
    Render.prototype.removeEntityById = function(id) {
        delete this.entities[id];
        this.changed = true;
    };
    Render.prototype.start = function(fps) {
        var interval = 1000 / fps; // calculate frames per second as interval
        var r = this;
        this.timer = setInterval(function() {
            r.draw();
        }, interval);
    };
    Render.prototype.draw = function() {
        if (!this.changed) { return; }
        this.changed = false;

        // It is very inefficent to be clearing the whole canvase each time
        // and we should not be referencing wsConn's statics. they should
        // be building objects we can injest instead of the messages.
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height)
        var _entities = this.entities;
        for (var id in _entities) {
            var e = _entities[id];
            if (e.T === WsConn.EntityTypes.block) {
                this.drawBlock(e);
            }
        }
        if (this.playersAdded) {
            this.players.sort(playerSort)
            this.playersAdded = false;
        }
        var _players = this.players;
        for (var idx = 0; idx < _players.length; idx++) {
            this.drawPlayer(_players[idx], idx+1)
        }
    };
    Render.prototype.drawBlock = function(block) {
        this.ctx.fillStyle = "rgba("+block.R+","+block.G+","+block.B+",0."+block.A+")";
        this.ctx.fillRect(block.X, block.Y, block.W, block.H);
    }
    Render.prototype.drawPlayer = function(player, slot) {
        this.ctx.fillStyle = "#000";
        this.ctx.font = "12pt Calibri";
        this.ctx.fillText(player.N + ": " + player.Sc, 0, 20 * slot);
    }


    return {
        runApp: runApp,
    };
})(this);
