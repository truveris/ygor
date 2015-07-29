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
        $scope.channelID = $routeParams.channelID;
        $scope.clientID = null;
        $scope.playlist = [];
        $scope.player = new Audio();
        $scope.playing = false;
        $scope.content = $("#ygor-content");


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

        $scope.playNext = function() {
            if ($scope.playlist.length > 0) {
                var item = $scope.playlist.shift()
                $scope.playing = true;
                $scope.player.src = item.URL;
                $scope.player.play();
                if (item.Duration !== null) {
                    setTimeout(function() { $scope.skip(); }, duration);
                }
            } else {
                $scope.playing = false;
            }
        }

        $scope.player.onended = function() {
            $scope.playNext();
        };

        $scope.volume = function(percent) {
            $scope.player.volume = parseInt(percent) / 100.0;
        }

        $scope.stop = function() {
            $scope.player.pause();
            $scope.playlist = [];
            $scope.playing = false;
        }

        $scope.skip = function() {
            $scope.player.pause();
            $scope.playNext();
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

            if (command.name == "volume") {
                $scope.volume(command.args[0]);
                return;
            }

            if (command.name == "xombrero open") {
                var url = command.args[0].replace(/http:/, "https:");
                $scope.content.html($("<iframe>").attr("src", url));
                return;
            }

            if (command.name == "play") {
                var url = command.args[0].replace(/http:/, "https:");
                var duration = null;
                if (command.args.length > 1) {
                    duration = parseFloat(command.args[1]);
                }
                $scope.playlist.push({"URL": url, "Duration": duration});
                if (!$scope.playing) {
                    $scope.playNext()
                }
            }
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
                            $scope.playlist = [];
                            $scope.showModal("disconnected");
                            break;
                    }
                })
                .error(function() {
                    $scope.showModal("disconnected");
                });
        }

        $scope.$on('$destroy', function() {
            $scope.clientID = null;
            $scope.player = null;
            $scope.content = null;
        });

        $scope.register = function() {
            $scope.showModal("connecting");

            $http.post("/channel/register", {"ChannelID": $scope.channelID})
                .success(function(data) {
                    $scope.clientID = data.ClientID;
                    $scope.showModal("waiting");
                    $scope.pollQueue();
                })
                .error(function() {
                    $scope.showModal("failed-register");
                });
        }

        $scope.register();
    }
]);
