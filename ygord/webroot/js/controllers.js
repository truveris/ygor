var ygorMinionControllers = angular.module('ygorMinionControllers', []);

ygorMinionControllers.controller("ChannelListController", ["$scope", "$http",
    function($scope, $http) {
        $http.get('/channel/list').success(function(data) {
            $scope.channels = data.channels;
        });
    }
]);

ygorMinionControllers.controller("AliasListController", ["$scope", "$http",
    function($scope, $http) {
        $http.get('/alias/list').success(function(data) {
            $scope.aliases = data.aliases;
        });
        $scope.orderProp = "Name";
    }
]);

ygorMinionControllers.controller("ClientListController", ["$scope", "$http",
    function($scope, $http) {
        $http.get('/client/list').success(function(data) {
            $scope.clients = data.clients;
        });
        $scope.orderProp = "Username";
    }
]);

ygorMinionControllers.controller("ChannelController", [
    "$scope", "$http", "$routeParams",
    function($scope, $http, $routeParams) {
        $(window).on("message", function(e) {
            $scope.handleChildMessage(e.originalEvent);
        });
        $scope.channelID = $routeParams.channelID;
        $scope.clientID = null;
        $scope.imageTrack = $("#ygor-content #imageTrack");
        $scope.playTrack = $("#ygor-content #playTrack");
        $scope.playTrack.playing = false;
        $scope.playTrack.playlist = [];
        $scope.content = $("#ygor-content");
        $scope.playTrack.css("visibility", "hidden");
        $scope.popUpContainer = $("#pop-up-container");
        var increment = 5;
        // set global volume variables for easy access by embedded iframes
        window.volume = 100;

        // imageTrack functions
        $scope.imageTrack.post = function(message) {
            $scope.imageTrack[0].contentWindow.postMessage(message, "*");
        }

        $scope.imageTrack.shutup = function() {
            $scope.imageTrack[0].contentWindow.shutup();
        }

        $scope.imageTrack.skip = function() {
            $scope.imageTrack[0].contentWindow.shutup();
        }

        $scope.imageTrack.stop = function() {
            $scope.imageTrack.shutup();
        }

        // playTrack functions
        $scope.playTrack.hide = function() {
            $scope.playTrack.css("visibility", "hidden");
        }

        $scope.playTrack.show = function() {
            $scope.playTrack.css("visibility", "visible");
        }

        $scope.playTrack.post = function(message) {
            $scope.playTrack[0].contentWindow.postMessage(message, "*");
        }

        $scope.playTrack.setVolume = function(level) {
            $scope.playTrack[0].contentWindow.setVolume(level);
        }

        $scope.playTrack.setTrackVolume = function(level) {
            window.qvolume = level;
            $scope.playTrack[0].contentWindow.setTrackVolume(level * 0.01);
        }

        $scope.playTrack.shutup = function() {
            $scope.playTrack.stop();
        }

        $scope.playTrack.skip = function() {
            $scope.playTrack[0].contentWindow.shutup();
        }

        $scope.playTrack.stop = function() {
            $scope.playTrack.playlist = [];
            $scope.playTrack[0].contentWindow.shutup();
        }

        // scope functions
        $scope.setVolume = function(level) {
            // set global volume variable for easy access by embedded iframes
            window.volume = level;
            //adjust track volumes
            $scope.playTrack.setVolume(level);
        }

        $scope.stop = function() {
            $scope.playTrack.stop();
        }

        $scope.skip = function() {
            $scope.playTrack.skip();
        }

        $scope.showError = function(srcTrack, submessage) {
            submessage = submessage || "";
            var $popUpDiv = $("<div>", {class: "error-pop-up"});
            var $titleP = $("<p>", {class: "error-title"});
            $titleP.html(srcTrack + " error");
            $popUpDiv.append($titleP)
            var $subtitleP = $("<p>", {class: "error-subtitle"});
            $subtitleP.html(submessage);
            $popUpDiv.append($subtitleP);
            $scope.popUpContainer.append($popUpDiv);
            // fade it out, then remove it.
            $popUpDiv.delay(2000).fadeOut({
                duration: 500,
                complete: function(){$(this).remove();}
            });
        }

        $(".button-reconnect").click(function() {
            $scope.register();
        });

        $scope.showModal = function(type) {
            $(".modal-options").hide();
            $("#modal-" + type).show();
            $("#modal").show();
        }

        $scope.hideModal = function() {
            $("#modal").hide();
        }

        $scope.playTrack.playNext = function() {
            if ($scope.playTrack.playlist.length > 0) {
                $scope.playTrack.playing = true;
                var media = $scope.playTrack.playlist.shift();
                $scope.playTrack.post(media);
            } else {
                $scope.playTrack.playing = false;
            }
        }

        $scope.handleChildMessage = function(event) {
            /* Ignore all messages that are not sent from the parent frame. */
            if (event.origin !== window.location.origin) {
                return;
            }

            var msg = event.data;
            var srcTrack = event.source.frameElement.id;
            switch (srcTrack){
                case "playTrack":
                    switch (msg.playerState) {
                        case "PLAYING":
                            $scope.playTrack.show();
                            break;
                        case "ENDED":
                            $scope.playTrack.hide();
                            $scope.playTrack.playing = false;
                            $scope.playTrack.playNext();
                            break;
                        case "ERRORED":
                            $scope.showError(srcTrack, msg.submessage);
                            break;
                    }
                    break;
                case "imageTrack":
                    switch (msg.playerState) {
                        case "ERRORED":
                            $scope.showError(srcTrack, msg.submessage);
                            break;
                    }
                    break;
            }
        }

        /*
         * handleCommand pushes a fresh command to the stack.  It also captures a
         * few special commands such as "skip" and "stop" and executes them
         * immediately.
         */
        $scope.handleCommand = function(command) {
            if (command.name == "skip") {
                $scope.skip();
                return;
            }

            if (command.name == "shutup") {
                $scope.stop();
                return;
            }

            if (command.name == "reboot") {
                document.location.reload();
                return;
            }

            if (command.name == "volume") {
                var level = command.data;
                // volume level must be between 100 and 0.0
                if (level == "1dB+") {
                    level = Math.min(100, volume + increment);
                } else if (level == "1dB-") {
                    level = Math.max(0, volume - increment);
                } else {
                    level = parseInt(level);
                    level = Math.max(0.0, Math.min(100, level));
                }
                $scope.setVolume(level);
                return;
            }

            if (command.name == "play") {
                $scope.playTrack.playlist.push(command.data);
                if (!$scope.playTrack.playing) {
                    $scope.playTrack.playNext()
                }
                return;
            }

            if (command.name == "image") {
                $scope.imageTrack.shutup();
                $scope.imageTrack.post(command.data);
                return;
            }
        }

        $scope.startReconnectCounter = function() {
            $scope.reconnectCounter = 10;
            $scope.reconnectInterval = setInterval(function() {
                $scope.reconnectCounter--;
                $(".modal-options span").text($scope.reconnectCounter);
                if ($scope.reconnectCounter <= 0) {
                    $scope.register();
                }
            }, 1000);
        }

        $scope.stopReconnectCounter = function() {
            clearInterval($scope.reconnectInterval);
        }

        /*
         * pollQueue runs for ever until it encounters a disconnection, it
         * feeds the internal playlist used by the command() function.
         */
        $scope.pollQueue = function() {
            if (!$scope.clientID)
                return;

            $http.post("/channel/poll", {"clientID": $scope.clientID})
                .success(function(data) {
                    switch (data.status) {
                        case "empty":
                            $scope.pollQueue();
                            break;
                        case "command":
                            $scope.hideModal();
                            for (var i = 0; i < data.commands.length; i++) {
                                $scope.handleCommand(data.commands[i]);
                            };
                            $scope.pollQueue();
                            break;
                        case "unknown-client":
                        default:
                            $scope.playTrack.playlist = [];
                            $scope.showModal("disconnected");
                            break;
                    }
                })
                .error(function() {
                    $scope.showModal("disconnected");
                    $scope.startReconnectCounter();
                });
        }

        $scope.$on('$destroy', function() {
            $scope.clientID = null;
            //$scope.player = null;
            $scope.content = null;
        });

        $scope.register = function() {
            $scope.stopReconnectCounter();
            $scope.showModal("connecting");

            $http.post("/channel/register", {"channelID": $scope.channelID})
                .success(function(data) {
                    $scope.clientID = data.clientID;
                    $scope.showModal("waiting");
                    $scope.pollQueue();
                })
                .error(function() {
                    $scope.showModal("failed-register");
                    $scope.startReconnectCounter();
                });
        }

        $scope.register();
    }
]);
