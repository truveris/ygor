var ygorMinionControllers = angular.module('ygorMinionControllers', []);

ygorMinionControllers.controller("ChannelListController", ["$scope", "$http",
    function($scope, $http) {
        $http.get('/channel/list').success(function(data) {
            $scope.channels = data.Channels;
        });
    }
]);

ygorMinionControllers.controller("AliasListController", ["$scope", "$http",
    function($scope, $http) {
        $http.get('/alias/list').success(function(data) {
            $scope.aliases = data.Aliases;
        });
        $scope.orderProp = "Name";
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
        $scope.musicTrack = $("#ygor-content #music-track");
        $scope.bgTrack = $("#ygor-content #bg-track");
        $scope.queueTrack = $("#ygor-content #queue-track");
        $scope.playTrack = $("#ygor-content #play-track");
        $scope.musicTrack.playing = false;
        $scope.queueTrack.playing = false;
        $scope.musicTrack.playlist = [];
        $scope.queueTrack.playlist = [];
        $scope.content = $("#ygor-content");
        $scope.musicTrack.attr("hidden", "hidden");
        $scope.queueTrack.attr("hidden", "hidden");
        $scope.playTrack.attr("hidden", "hidden");
        var increment = 0.05;
        // set global volume variable for easy access by embedded iframes
        window.volume = 1.0;

        // musicTrack functions
        $scope.musicTrack.hide = function() {
            $scope.musicTrack.attr("hidden", "hidden");
        }

        $scope.musicTrack.show = function() {
            $scope.musicTrack.removeAttr("hidden");
        }

        $scope.musicTrack.post = function(message) {
            $scope.musicTrack[0].contentWindow.postMessage(JSON.stringify(message), "*");
        }

        $scope.musicTrack.setVolume = function(level) {
            $scope.musicTrack[0].contentWindow.setVolume(level);
        }

        $scope.musicTrack.shutup = function() {
            $scope.musicTrack[0].contentWindow.shutup();
        }

        $scope.musicTrack.skip = function() {
            $scope.musicTrack.shutup();
        }

        $scope.musicTrack.stop = function() {
            $scope.musicTrack.playlist = [];
            $scope.musicTrack.shutup();
        }

        // bgTrack functions
        $scope.bgTrack.post = function(message) {
            $scope.bgTrack[0].contentWindow.postMessage(JSON.stringify(message), "*");
        }

        $scope.bgTrack.shutup = function() {
            $scope.bgTrack[0].contentWindow.shutup();
        }

        $scope.bgTrack.skip = function() {
            $scope.bgTrack.shutup();
        }

        $scope.bgTrack.stop = function() {
            $scope.bgTrack.shutup();
        }

        // queueTrack functions
        $scope.queueTrack.hide = function() {
            $scope.queueTrack.attr("hidden", "hidden");
        }

        $scope.queueTrack.show = function() {
            $scope.queueTrack.removeAttr("hidden");
        }

        $scope.queueTrack.post = function(message) {
            $scope.queueTrack[0].contentWindow.postMessage(JSON.stringify(message), "*");
        }

        $scope.queueTrack.setVolume = function(level) {
            $scope.queueTrack[0].contentWindow.setVolume(level);
        }

        $scope.queueTrack.shutup = function() {
            $scope.queueTrack[0].contentWindow.shutup();
        }

        $scope.queueTrack.skip = function() {
            $scope.queueTrack.shutup();
        }

        $scope.queueTrack.stop = function() {
            $scope.queueTrack.playlist = [];
            $scope.queueTrack.shutup();
        }

        // playTrack functions
        $scope.playTrack.hide = function() {
            $scope.playTrack.attr("hidden", "hidden");
        }

        $scope.playTrack.show = function() {
            $scope.playTrack.removeAttr("hidden");
        }

        $scope.playTrack.post = function(message) {
            $scope.playTrack[0].contentWindow.postMessage(JSON.stringify(message), "*");
        }

        $scope.playTrack.setVolume = function(level) {
            $scope.playTrack[0].contentWindow.setVolume(level);
        }

        $scope.playTrack.shutup = function() {
            $scope.playTrack[0].contentWindow.shutup();
        }

        $scope.playTrack.skip = function() {
            $scope.playTrack.shutup();
        }

        $scope.playTrack.stop = function() {
            $scope.playTrack.shutup();
        }

        // scope functions
        $scope.setVolume = function(level) {
            // set global volume variable for easy access by embedded iframes
            window.volume = level;
            //adjust track volumes
            $scope.queueTrack.setVolume(level);
            $scope.musicTrack.setVolume(level);
            $scope.playTrack.setVolume(level);
        }

        $scope.stop = function() {
            $scope.queueTrack.stop();
            $scope.musicTrack.stop();
            $scope.playTrack.stop();
        }

        $scope.skip = function() {
            $scope.queueTrack.skip();
            $scope.musicTrack.skip();
        }

        $scope.showError = function(srcTrack, submessage) {
            submessage = submessage || "";
            var popUp = document.createElement("div");
            popUp.setAttribute("class", "error-pop-up");
            var title = document.createElement("p");
            title.setAttribute("class", "error-title");
            title.innerHTML = srcTrack + " error";
            popUp.appendChild(title);
            var subtitle = document.createElement("p");
            subtitle.setAttribute("class", "error-subtitle");
            subtitle.innerHTML = submessage;
            popUp.appendChild(subtitle);
            document.getElementById("pop-up-container").appendChild(popUp);
            popUp.style.opacity = 1;
            // Make sure the initial state is applied.
            window.getComputedStyle(popUp).opacity;
            popUp.style.opacity = 0;
            setTimeout(function() {
                popUp.parentNode.removeChild(popUp);
            }, 4000);
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

        $scope.musicTrack.playNext = function() {
            if ($scope.musicTrack.playlist.length > 0) {
                var mediaObj = $scope.musicTrack.playlist.shift();
                $scope.musicTrack.playing = true;
                $scope.musicTrack.post(mediaObj);
            } else {
                $scope.musicTrack.playing = false;
            }
        }

        $scope.queueTrack.playNext = function() {
            if ($scope.queueTrack.playlist.length > 0) {
                var mediaObj = $scope.queueTrack.playlist.shift();
                $scope.queueTrack.playing = true;
                $scope.queueTrack.post(mediaObj);
            } else {
                $scope.queueTrack.playing = false;
            }
        }

        $scope.handleChildMessage = function(event) {
            if (event.origin !== "http://localhost:8181" &&
                event.origin !== "https://truveris.com"){
                return;
            }
            msg = JSON.parse(event.data);
            switch (msg.source){
                case "queueTrack":
                    switch (msg.playerState) {
                        case "PLAYING":
                            $scope.queueTrack.show();
                            break;
                        case "ENDED":
                            $scope.queueTrack.hide();
                            $scope.queueTrack.playing = false;
                            $scope.queueTrack.playNext();
                            break;
                        case "ERRORED":
                            $scope.showError(msg.source, msg.submessage);
                            $scope.queueTrack.hide();
                            $scope.queueTrack.shutup();
                            $scope.queueTrack.playing = false;
                            $scope.queueTrack.playNext();
                            break;
                    }
                    break;
                case "musicTrack":
                    switch (msg.playerState) {
                        case "ENDED":
                            $scope.musicTrack.playing = false;
                            $scope.musicTrack.playNext();
                            break;
                        case "ERRORED":
                            $scope.showError(msg.source, msg.submessage);
                            $scope.musicTrack.hide();
                            $scope.musicTrack.shutup();
                            $scope.musicTrack.playing = false;
                            $scope.musicTrack.playNext();
                            break;
                    }
                    break;
                case "playTrack":
                    switch (msg.playerState) {
                        case "PLAYING":
                            $scope.playTrack.show();
                            break;
                        case "ENDED":
                            $scope.playTrack.hide();
                            break;
                        case "ERRORED":
                            $scope.showError(msg.source, msg.submessage);
                            break;
                    }
                    break;
                case "bgTrack":
                    switch (msg.playerState) {
                        case "ERRORED":
                            $scope.showError(msg.source, msg.submessage);
                            break;
                    }
                    break;
            }
        }

        /*
         * translateCommand will convert a single string into a command object,
         * parsing out the useful information.  At some point in the future we
         * will receive commands pre-parsed, but not until the web minions are
         * predominant.
         */
        $scope.translateCommand = function(msg) {
            var command = {};

            var tokens = msg.split(" ");
            if (tokens[0] == "xombrero") {
                command.name = tokens[0] + " " + tokens[1];
                tokens.shift();
                tokens.shift();
            } else {
                command.name = tokens[0];
                tokens.shift();
            }
            command.args = tokens

            return command;
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
                var level = command.args[0];
                // volume level must be between 1.0 and 0.0
                if (level == "1dB+") {
                    level = Math.min(1.0, volume + increment);
                } else if (level == "1dB-") {
                    level = Math.max(0.0, volume - increment);
                } else {
                    level = parseInt(level) / 100.0;
                    level = Math.max(0.0, Math.min(1.0, level));
                }
                $scope.setVolume(level);
                return;
            }

            if (command.name == "xombrero open") {
                // var url = command.args[0].replace(/http:/, "https:");
                var url = command.args[0];
                $scope.content.html($("<iframe>").attr("src", url));
                return;
            }

            if (command.name == "queue") {
                mediaObj = JSON.parse(command.args[0]);
                $scope.queueTrack.playlist.push(mediaObj);
                if (!$scope.queueTrack.playing) {
                    $scope.queueTrack.playNext()
                }
                return;
            }

            if (command.name == "music") {
                mediaObj = JSON.parse(command.args[0]);
                $scope.musicTrack.playlist.push(mediaObj);
                if (!$scope.musicTrack.playing) {
                    $scope.musicTrack.playNext()
                }
                return;
            }

            if (command.name == "bg") {
                mediaObj = JSON.parse(command.args[0]);
                $scope.bgTrack.shutup();
                $scope.bgTrack.post(mediaObj);
                return;
            }

            if (command.name == "play") {
                mediaObj = JSON.parse(command.args[0]);
                $scope.playTrack.post(mediaObj);
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

            $http.post("/channel/poll", {"ClientID": $scope.clientID})
                .success(function(data) {
                    switch (data.Status) {
                        case "empty":
                            $scope.pollQueue();
                            break;
                        case "command":
                            $scope.hideModal();
                            for (var i = 0; i < data.Commands.length; i++) {
                                cmd = $scope.translateCommand(data.Commands[i]);
                                $scope.handleCommand(cmd);
                            };
                            $scope.pollQueue();
                            break;
                        case "unknown-client":
                        default:
                            $scope.queueTrack.playlist = [];
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

            $http.post("/channel/register", {"ChannelID": $scope.channelID})
                .success(function(data) {
                    $scope.clientID = data.ClientID;
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
