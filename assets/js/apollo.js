
var ApolloApp = (function(context){
    var config = null;
    var ws = null;
    var board = null;

    function selected(id) {
        if (!ws.conn) {
            return;
        }

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

        var wnd = $(window)
        board = newGameBoard(cfg.container, wnd.width(), wnd.height())
        wnd.resize(function() {
            if(this.resizeTO) clearTimeout(this.resizeTO);
            this.resizeTO = setTimeout(function() {
                wnd.trigger('resizeEnd');
            }, 500);
        });
        wnd.bind('resizeEnd', function() {
            board.resize(wnd.width(), wnd.height());
        })

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
            var gameType = msg.Gt;
            if (gameType) {
                board.setGameType(gameType)
            }
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
                    this.board.updateEntity(entity);
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
                    this.board.updatePlayer(player);
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

    function newGameBoard(container, startWidth, startHeight) {
        var entityColors = ['red', 'blue', 'green', 'gray', 'orange'];
        var contNode = document.getElementById(container);
        var stage = null,
            playerLayer = null,
            entLayer = null,
            players  = [],
            entities = [],
            gridInfo = {
                width: startWidth, height: startHeight,
                rStep: 0, cStep: 0,
                rHalf: 0, cHalf: 0,
                rPad: 0,  cPad: 0,
            },
            entGrid = [],
            gameType = {
                rows: 0, cols: 0
            },
            selcColor = null;

        function calulateGrid(width, height) {
            width -= 10; height -= 10;
            rStep = height/gameType.rows;
            cStep = width/gameType.cols;
            gridInfo = {
                width: width, height: height,
                rStep: rStep, cStep: cStep,
                rHalf: rStep/2, cHalf: cStep/2,
                rPad: 10, cPad: 10
            };
        }

        function updateEntForSelect(state, drawable) {
            var cPad = gridInfo.cPad,
                rPad = gridInfo.rPad,
                alpha = 1;

            if (state === WsConn.EntityUpdateTypes.selected) {
                alpha = 0.5;
                cPad = gridInfo.cHalf * 0.5;
                rPad = gridInfo.rHalf * 0.5;
            }

            drawable.setAlpha(alpha);
            drawable.setCenterOffset(-1*cPad, -1*rPad);
            drawable.setSize(gridInfo.cStep - (cPad*2), gridInfo.rStep - (rPad*2));
        }

        // Players
        function findPlayer(id) {
            var pLen =  players.length;
            var idx;
            for (idx=pLen-1; idx >= 0; idx--) {
                var player = players[idx];
                if (player.p.Id === id) {
                    break;
                }
            }
            return idx;
        }

        function addPlayer(player) {
            var idx = findPlayer(player.Id);
            if (idx !== -1) { return; }

            var text = new Kinetic.Text({
                x: 0,
                y: players.length * 24,
                fontSize: 22,
                textFill: 'black',
                fontFamily: "Calibri",
                text: 'P '+player.Id+' score: '+player.Sc,
            });
            playerLayer.add(text);
            playerLayer.draw();

            players.push({p: player, d: text});
        }

        function removePlayer(id) {
            var idx = findPlayer(id);
            if (idx === -1) {
                console.error('tried to removed a player I dont know about;', id);
                return
            }

            playerObj = players[idx];
            playerObj.d.clearData();
            playerLayer.remove(playerObj.d);
            playerObj.d = null;

            players.splice(idx, 1);

            // Update indexes to reflect removal
            for (idx=0; idx<players.length; idx++) {
                players[idx].d.setY(idx * 24);
            }
            playerLayer.draw();
        }

        function updatePlayer(player) {
            var idx = findPlayer(player.Id);
            if (idx !== -1) {
                playerCont = players[idx];
                playerCont.p = player;
                playerCont.d.setText('P '+player.Id+' score: '+player.Sc);
                playerLayer.draw();
            } else {
                addPlayer(player)
                console.info('tried to update a player I dont know about, adding;', player);
            }
        }

        // Entities
        function addEntity(entity) {
            // Inore duplicate items
            if (entities[entity.Id]) { return; }

            entity.color = entityColors[entity.C];
            var rect = new Kinetic.Rect({
                x: (gridInfo.cStep * entity.X),
                y: (gridInfo.rStep * entity.Y),
                fill:   entity.color,
                stroke: "black",
                strokeWidth: 2,
            });
            updateEntForSelect(entity.St, rect);

            var entObj = {e: entity, d: rect};
            if (!entGrid[entity.X]) { entGrid[entity.X] = []; }
            entGrid[entity.X][entity.Y] = entObj;
            entities[entity.Id] = entObj;

            entLayer.add(rect)
            entLayer.draw();
        }

        function updateEntity(entity) {
            var entObj = entities[entity.Id];
            if (!entObj) {
                addEntity(entity);
                return;
            }
            entObj.e = entity;

            updateEntForSelect(entity.St, entObj.d);
            entLayer.draw();
        }

        function removeEntity(id) {
            if (!entities[id]) {
                return;
            }

            entObj = entities[id];
            entGrid[entObj.e.X][entObj.e.Y] = null;

            entObj.d.clearData();
            entLayer.remove(entObj.d);
            entObj.d = null;
            delete entities[id];
            entLayer.draw();
        }

        function resize(width, height) {
            calulateGrid(width, height);
            stage.setSize(width, height);

            for (var id in entities) {
                if (!entities.hasOwnProperty(id)) { continue; }
                var entObj = entities[id];
                entObj.d.setX(gridInfo.cStep * entObj.e.X);
                entObj.d.setY(gridInfo.rStep * entObj.e.Y);

                updateEntForSelect(entObj.e.St, entObj.d);
            }
            entLayer.draw();
        }

        function setGameType(gt) {
            gameType.rows = gt.R;
            gameType.cols = gt.C;
            calulateGrid(gridInfo.width, gridInfo.height)
        }

        // Get the coordinates for a mouse or touch event
        function getCoords(e) {
            if (e.clientX) {
                return { x: e.clientX, y: e.clientY };
            }
            else if (e.offsetX) {
                return { x: e.offsetX, y: e.offsetY };
            }
            else if (e.layerX) {
                return { x: e.layerX, y: e.layerY };
            }
            else {
                return { x: e.pageX - contNode.offsetLeft, y: e.pageY - contNode.offsetTop };
            }
        }

        function init() {
            calulateGrid(gridInfo.width, gridInfo.height);

            stage = new Kinetic.Stage({
                container: contNode,
                width: gridInfo.width, height: gridInfo.height
            });

            playerLayer = new Kinetic.Layer();
            entLayer = new Kinetic.Layer();

            stage.add(entLayer);
            stage.add(playerLayer);

            $('#'+container).bind('mousedown touchstart', function(e) {
                var point = null;
                if(e.originalEvent.touches && e.originalEvent.touches.length) {
                    e = e.originalEvent.touches[0];
                } else if(e.originalEvent.changedTouches && e.originalEvent.changedTouches.length) {
                    e = e.originalEvent.changedTouches[0];
                }
                point = getCoords(e);
                var x = Math.floor(point.x / gridInfo.cStep);
                var y = Math.floor(point.y / gridInfo.rStep);
                if (entGrid[x] && entGrid[x][y]) {
                    var entObj =  entGrid[x][y];

                    if (entObj.e.St === WsConn.EntityUpdateTypes.selected) {
                        entObj.e.St = WsConn.EntityUpdateTypes.present;
                        selected(entObj.e.Id);

                    } else {
                        // TODO make it so we never send the select msg out
                        // if the color is not matching.
                        entObj.e.St = WsConn.EntityUpdateTypes.selected;
                        selected(entObj.e.Id);
                    }

                    updateEntForSelect(entObj.e.St, entObj.d);
                    entLayer.draw();
                }
            });
        }


        // Initialize the game board
        init();


        return {
            setGameType:  setGameType,
            addPlayer:    addPlayer,
            removePlayer: removePlayer,
            updatePlayer: updatePlayer,
            addEntity:    addEntity,
            updateEntity: updateEntity,
            removeEntity: removeEntity,
            resize:       resize,
        }
    }

    return {
        runApp: runApp,
    };
})(this);
