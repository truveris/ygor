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
        $scope.queueID = null;
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
                $scope.playing = true;
                $scope.player.src = $scope.playlist.shift();
                $scope.player.play();
            } else {
                $scope.playing = false;
            }
        }

        $scope.player.onended = function() {
            $scope.playNext();
        };

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
            command.args = tokens.join(" ");

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

            if (command.name == "xombrero open") {
                $scope.content.html($("<iframe>").attr("src", command.args));
                return;
            }

            if (command.name == "play") {
                $scope.playlist.push(command.args);
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
            if (!$scope.queueID)
                return;

            $http.post("/channel/poll", {"QueueID": $scope.queueID})
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
                        case "queue-not-found":
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
            $scope.queueID = null;
            $scope.player = null;
            $scope.content = null;
        });

        $scope.register = function() {
            $scope.showModal("connecting");

            $http.post("/channel/register", {"ChannelID": $scope.channelID})
                .success(function(data) {
                    $scope.queueID = data.QueueID;
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
